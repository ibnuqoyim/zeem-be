package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"zeem/internal/models"
	"zeem/internal/services"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// SignalingMessage represents a WebRTC signaling message
type SignalingMessage struct {
	Type     string      `json:"type"`
	RoomID   string      `json:"roomId"`
	SenderID string      `json:"senderId"`
	Data     interface{} `json:"data"`
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	roomManager   *services.RoomManager
	webrtcManager *services.WebRTCManager
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(rm *services.RoomManager, wm *services.WebRTCManager) *WebSocketHandler {
	return &WebSocketHandler{
		roomManager:   rm,
		webrtcManager: wm,
	}
}

// HandleConnection handles incoming WebSocket connections
func (h *WebSocketHandler) HandleConnection(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	participantID := uuid.New().String()
	roomID := c.Query("roomId")
	roomType := models.ConnectionType(c.Query("type"))
	username := c.Query("username")
	isBroadcaster := c.Query("broadcaster") == "true"
	isScreenShare := c.Query("screenShare") == "true"

	if roomID == "" {
		log.Println("Room ID not provided")
		return
	}

	// Validate connection type
	switch roomType {
	case models.OneToOne, models.Broadcasting, models.ScreenSharing:
		// Valid types
	default:
		roomType = models.OneToOne // Default to one-to-one
	}

	// Get or create room
	var room *models.Room
	if !h.roomManager.RoomExists(roomID) {
		room = h.roomManager.CreateRoom(roomID, roomType)
	} else {
		room = h.roomManager.GetRoom(roomID)
	}

	// Create participant
	participant := &models.Participant{
		ID:       participantID,
		Conn:     conn,
		Username: username,
		ConnectionInfo: &models.ConnectionInfo{
			Type:          roomType,
			IsBroadcaster: isBroadcaster,
			IsScreenShare: isScreenShare,
		},
	}

	// Try to add participant
	if err := room.AddParticipant(participant); err != nil {
		log.Printf("Failed to add participant: %v", err)
		conn.WriteJSON(SignalingMessage{
			Type: "error",
			Data: err.Error(),
		})
		return
	}

	defer func() {
		room.RemoveParticipant(participantID)
		h.webrtcManager.RemovePeerConnection(participantID)

		// Notify others about participant leaving
		h.broadcastToRoom(room, SignalingMessage{
			Type:     "participant_left",
			RoomID:   roomID,
			SenderID: participantID,
		}, participantID)
	}()

	// Send room info to the new participant
	conn.WriteJSON(SignalingMessage{
		Type: "room_info",
		Data: map[string]interface{}{
			"roomId":       roomID,
			"roomType":     roomType,
			"participants": room.GetParticipants(),
			"chatHistory":  room.GetChatHistory(),
		},
	})

	// Notify others about new participant
	h.broadcastToRoom(room, SignalingMessage{
		Type:     "participant_joined",
		RoomID:   roomID,
		SenderID: participantID,
		Data: map[string]interface{}{
			"username":      username,
			"isBroadcaster": isBroadcaster,
			"isScreenShare": isScreenShare,
		},
	}, participantID)

	// Handle messages
	for {
		var msg SignalingMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		msg.SenderID = participantID
		msg.RoomID = roomID

		// Handle different message types
		switch msg.Type {
		case "offer":
			// Forward offer to the target participant
			h.broadcastToRoom(room, msg, participantID)

		case "answer":
			// Forward answer to the target participant
			h.broadcastToRoom(room, msg, participantID)

		case "ice_candidate":
			// Forward ICE candidate to the target participant
			h.broadcastToRoom(room, msg, participantID)

		case "chat":
			if content, ok := msg.Data.(string); ok {
				chatMsg := models.ChatMessage{
					SenderID:   participantID,
					SenderName: participant.Username,
					Content:    content,
					Timestamp:  time.Now().Unix(),
				}
				room.AddChatMessage(chatMsg)
				h.broadcastToRoom(room, SignalingMessage{
					Type:     "chat",
					RoomID:   roomID,
					SenderID: participantID,
					Data:     chatMsg,
				}, "")
			}

		case "screen_share_start":
			participant.ConnectionInfo.IsScreenShare = true
			h.broadcastToRoom(room, msg, participantID)

		case "screen_share_stop":
			participant.ConnectionInfo.IsScreenShare = false
			h.broadcastToRoom(room, msg, participantID)

		default:
			// Broadcast other messages to room participants
			h.broadcastToRoom(room, msg, participantID)
		}
	}
}

// broadcastToRoom sends a message to all participants in a room except the sender
func (h *WebSocketHandler) broadcastToRoom(room *models.Room, msg SignalingMessage, excludeID string) {
	participants := room.GetParticipants()
	for _, p := range participants {
		if excludeID == "" || p.ID != excludeID {
			err := p.Conn.WriteJSON(msg)
			if err != nil {
				log.Printf("Error sending message to participant %s: %v", p.ID, err)
			}
		}
	}
}
