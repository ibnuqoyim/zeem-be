package services

import (
	"testing"

	"zeem/internal/models"
)

func TestRoomManager(t *testing.T) {
	rm := NewRoomManager()

	// Test CreateRoom
	roomID := "test-room"
	roomType := models.OneToOne
	room := rm.CreateRoom(roomID, roomType)

	if room.ID != roomID {
		t.Errorf("Expected room ID %s, got %s", roomID, room.ID)
	}

	if room.Type != roomType {
		t.Errorf("Expected room type %s, got %s", roomType, room.Type)
	}

	// Test RoomExists
	if !rm.RoomExists(roomID) {
		t.Error("Expected room to exist")
	}

	if rm.RoomExists("non-existent") {
		t.Error("Expected non-existent room to not exist")
	}

	// Test GetRoom
	retrievedRoom := rm.GetRoom(roomID)
	if retrievedRoom != room {
		t.Error("Expected to get the same room instance")
	}

	// Test DeleteRoom
	rm.DeleteRoom(roomID)
	if rm.RoomExists(roomID) {
		t.Error("Expected room to be deleted")
	}
}

func TestGetRoomsByType(t *testing.T) {
	rm := NewRoomManager()

	// Create rooms of different types
	oneToOneRoom := rm.CreateRoom("one-to-one", models.OneToOne)
	broadcastRoom1 := rm.CreateRoom("broadcast-1", models.Broadcasting)
	broadcastRoom2 := rm.CreateRoom("broadcast-2", models.Broadcasting)
	screenShareRoom := rm.CreateRoom("screen-share", models.ScreenSharing)

	// Test getting OneToOne rooms
	oneToOneRooms := rm.GetRoomsByType(models.OneToOne)
	if len(oneToOneRooms) != 1 {
		t.Errorf("Expected 1 OneToOne room, got %d", len(oneToOneRooms))
	}
	if oneToOneRooms[0] != oneToOneRoom {
		t.Error("Expected to get the same OneToOne room instance")
	}

	// Test getting Broadcasting rooms
	broadcastRooms := rm.GetRoomsByType(models.Broadcasting)
	if len(broadcastRooms) != 2 {
		t.Errorf("Expected 2 Broadcasting rooms, got %d", len(broadcastRooms))
	}
	foundBroadcast1 := false
	foundBroadcast2 := false
	for _, room := range broadcastRooms {
		if room == broadcastRoom1 {
			foundBroadcast1 = true
		}
		if room == broadcastRoom2 {
			foundBroadcast2 = true
		}
	}
	if !foundBroadcast1 || !foundBroadcast2 {
		t.Error("Expected to find both broadcast rooms")
	}

	// Test getting ScreenSharing rooms
	screenShareRooms := rm.GetRoomsByType(models.ScreenSharing)
	if len(screenShareRooms) != 1 {
		t.Errorf("Expected 1 ScreenSharing room, got %d", len(screenShareRooms))
	}
	if screenShareRooms[0] != screenShareRoom {
		t.Error("Expected to get the same ScreenSharing room instance")
	}
}
