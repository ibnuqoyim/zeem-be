package models

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Participant represents a user in a room
type Participant struct {
	ID             string
	Conn           *websocket.Conn
	Username       string
	ConnectionInfo *ConnectionInfo
}

// Room represents a video conference room
type Room struct {
	ID           string
	Type         ConnectionType
	Participants map[string]*Participant
	Broadcaster  *Participant // For broadcasting mode
	mutex        sync.RWMutex
	ChatHistory  []ChatMessage
}

// ChatMessage represents a chat message in the room
type ChatMessage struct {
	SenderID   string `json:"sender_id"`
	SenderName string `json:"sender_name"`
	Content    string `json:"content"`
	Timestamp  int64  `json:"timestamp"`
}

// NewRoom creates a new room instance
func NewRoom(id string, roomType ConnectionType) *Room {
	return &Room{
		ID:           id,
		Type:         roomType,
		Participants: make(map[string]*Participant),
		ChatHistory:  make([]ChatMessage, 0),
	}
}

// AddParticipant adds a new participant to the room
func (r *Room) AddParticipant(p *Participant) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	switch r.Type {
	case OneToOne:
		if len(r.Participants) >= 2 {
			return ErrRoomFull
		}
	case Broadcasting:
		if p.ConnectionInfo.IsBroadcaster {
			if r.Broadcaster != nil {
				return ErrBroadcasterExists
			}
			r.Broadcaster = p
		}
	}

	r.Participants[p.ID] = p
	return nil
}

// RemoveParticipant removes a participant from the room
func (r *Room) RemoveParticipant(participantID string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if p := r.Participants[participantID]; p != nil {
		if r.Broadcaster != nil && r.Broadcaster.ID == participantID {
			r.Broadcaster = nil
		}
		delete(r.Participants, participantID)
	}
}

// GetParticipant gets a participant by ID
func (r *Room) GetParticipant(participantID string) *Participant {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.Participants[participantID]
}

// GetParticipants returns all participants in the room
func (r *Room) GetParticipants() []*Participant {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	participants := make([]*Participant, 0, len(r.Participants))
	for _, p := range r.Participants {
		participants = append(participants, p)
	}
	return participants
}

// AddChatMessage adds a new chat message to the room history
func (r *Room) AddChatMessage(message ChatMessage) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.ChatHistory = append(r.ChatHistory, message)
}

// GetChatHistory returns the chat history
func (r *Room) GetChatHistory() []ChatMessage {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.ChatHistory
}
