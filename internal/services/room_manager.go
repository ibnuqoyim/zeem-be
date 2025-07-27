package services

import (
	"sync"

	"zeem/internal/models"
)

// RoomManager handles the management of video conference rooms
type RoomManager struct {
	rooms map[string]*models.Room
	mutex sync.RWMutex
}

// NewRoomManager creates a new instance of RoomManager
func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*models.Room),
	}
}

// CreateRoom creates a new room with the given ID and type
func (rm *RoomManager) CreateRoom(roomID string, roomType models.ConnectionType) *models.Room {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	room := models.NewRoom(roomID, roomType)
	rm.rooms[roomID] = room
	return room
}

// GetRoom returns a room by its ID
func (rm *RoomManager) GetRoom(roomID string) *models.Room {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	return rm.rooms[roomID]
}

// DeleteRoom removes a room
func (rm *RoomManager) DeleteRoom(roomID string) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	delete(rm.rooms, roomID)
}

// RoomExists checks if a room exists
func (rm *RoomManager) RoomExists(roomID string) bool {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	_, exists := rm.rooms[roomID]
	return exists
}

// GetRoomsByType returns all rooms of a specific type
func (rm *RoomManager) GetRoomsByType(roomType models.ConnectionType) []*models.Room {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	var rooms []*models.Room
	for _, room := range rm.rooms {
		if room.Type == roomType {
			rooms = append(rooms, room)
		}
	}
	return rooms
}
