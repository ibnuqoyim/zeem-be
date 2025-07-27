package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"zeem/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func setupTestServer() (*gin.Engine, *services.RoomManager, *services.WebRTCManager) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	roomManager := services.NewRoomManager()
	webrtcManager := services.NewWebRTCManager()
	wsHandler := NewWebSocketHandler(roomManager, webrtcManager)

	router.GET("/ws", wsHandler.HandleConnection)
	return router, roomManager, webrtcManager
}

func createTestWebSocketConnection(t *testing.T, server *gin.Engine, query string) *websocket.Conn {
	s := httptest.NewServer(server)
	defer s.Close()

	// Convert http to ws
	wsURL := "ws" + strings.TrimPrefix(s.URL, "http") + "/ws" + query

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("could not open websocket connection: %v", err)
	}

	return ws
}

func readMessage(ws *websocket.Conn, timeout time.Duration) (*SignalingMessage, error) {
	var msg SignalingMessage
	ws.SetReadDeadline(time.Now().Add(timeout))
	err := ws.ReadJSON(&msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func waitForMessage(t *testing.T, ws *websocket.Conn, expectedType string) *SignalingMessage {
	for i := 0; i < 3; i++ { // Try up to 3 times
		msg, err := readMessage(ws, 1*time.Second)
		if err != nil {
			continue
		}
		if msg.Type == expectedType {
			return msg
		}
	}
	t.Fatalf("did not receive expected message type: %s", expectedType)
	return nil
}

func TestWebSocketHandler_OneToOne(t *testing.T) {
	router, _, _ := setupTestServer()

	// Create two WebSocket connections
	ws1 := createTestWebSocketConnection(t, router, "?roomId=test-room&type=one_to_one&username=user1")
	defer ws1.Close()

	// Wait for room info message
	msg1 := waitForMessage(t, ws1, "room_info")
	if msg1 == nil {
		t.Fatal("did not receive room info message")
	}

	ws2 := createTestWebSocketConnection(t, router, "?roomId=test-room&type=one_to_one&username=user2")
	defer ws2.Close()

	// Wait for room info message on second connection
	msg2 := waitForMessage(t, ws2, "room_info")
	if msg2 == nil {
		t.Fatal("did not receive room info message")
	}

	// Test chat message
	chatMsg := SignalingMessage{
		Type:   "chat",
		RoomID: "test-room",
		Data:   "Hello, World!",
	}
	err := ws1.WriteJSON(chatMsg)
	if err != nil {
		t.Fatalf("could not send chat message: %v", err)
	}

	// Wait for chat message on second connection
	receivedMsg := waitForMessage(t, ws2, "chat")
	if receivedMsg == nil {
		t.Fatal("did not receive chat message")
	}
}

func TestWebSocketHandler_Broadcasting(t *testing.T) {
	router, _, _ := setupTestServer()

	// Create broadcaster connection
	broadcaster := createTestWebSocketConnection(t, router, "?roomId=test-room&type=broadcasting&username=broadcaster&broadcaster=true")
	defer broadcaster.Close()

	// Wait for room info message
	msg := waitForMessage(t, broadcaster, "room_info")
	if msg == nil {
		t.Fatal("did not receive room info message")
	}

	// Create viewer connections
	viewer1 := createTestWebSocketConnection(t, router, "?roomId=test-room&type=broadcasting&username=viewer1")
	defer viewer1.Close()

	viewer2 := createTestWebSocketConnection(t, router, "?roomId=test-room&type=broadcasting&username=viewer2")
	defer viewer2.Close()

	// Wait for room info messages on viewers
	waitForMessage(t, viewer1, "room_info")
	waitForMessage(t, viewer2, "room_info")

	// Test broadcasting message
	broadcastMsg := SignalingMessage{
		Type:   "broadcast",
		RoomID: "test-room",
		Data:   "Broadcast message",
	}
	err := broadcaster.WriteJSON(broadcastMsg)
	if err != nil {
		t.Fatalf("could not send broadcast message: %v", err)
	}

	// Wait for broadcast messages on viewers
	msg1 := waitForMessage(t, viewer1, "broadcast")
	msg2 := waitForMessage(t, viewer2, "broadcast")

	if msg1 == nil || msg2 == nil {
		t.Fatal("viewers did not receive broadcast message")
	}
}

func TestWebSocketHandler_ScreenSharing(t *testing.T) {
	router, _, _ := setupTestServer()

	// Create screen sharer connection
	sharer := createTestWebSocketConnection(t, router, "?roomId=test-room&type=screen_sharing&username=sharer&screenShare=true")
	defer sharer.Close()

	// Wait for room info message
	waitForMessage(t, sharer, "room_info")

	// Create viewer connection
	viewer := createTestWebSocketConnection(t, router, "?roomId=test-room&type=screen_sharing&username=viewer")
	defer viewer.Close()

	// Wait for room info message
	waitForMessage(t, viewer, "room_info")

	// Test screen share start
	startMsg := SignalingMessage{
		Type:   "screen_share_start",
		RoomID: "test-room",
		Data:   "Screen share started",
	}
	err := sharer.WriteJSON(startMsg)
	if err != nil {
		t.Fatalf("could not send screen share start message: %v", err)
	}

	// Wait for screen share start message
	msg1 := waitForMessage(t, viewer, "screen_share_start")
	if msg1 == nil {
		t.Fatal("did not receive screen share start message")
	}

	// Test screen share stop
	stopMsg := SignalingMessage{
		Type:   "screen_share_stop",
		RoomID: "test-room",
		Data:   "Screen share stopped",
	}
	err = sharer.WriteJSON(stopMsg)
	if err != nil {
		t.Fatalf("could not send screen share stop message: %v", err)
	}

	// Wait for screen share stop message
	msg2 := waitForMessage(t, viewer, "screen_share_stop")
	if msg2 == nil {
		t.Fatal("did not receive screen share stop message")
	}
}
