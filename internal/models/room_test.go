package models

import (
	"testing"
)

func TestNewRoom(t *testing.T) {
	roomID := "test-room"
	roomType := OneToOne

	room := NewRoom(roomID, roomType)

	if room.ID != roomID {
		t.Errorf("Expected room ID %s, got %s", roomID, room.ID)
	}

	if room.Type != roomType {
		t.Errorf("Expected room type %s, got %s", roomType, room.Type)
	}

	if len(room.Participants) != 0 {
		t.Errorf("Expected empty participants map, got %d participants", len(room.Participants))
	}
}

func TestAddParticipant(t *testing.T) {
	tests := []struct {
		name          string
		roomType      ConnectionType
		participants  []*Participant
		expectError   bool
		errorExpected error
	}{
		{
			name:     "OneToOne_Success",
			roomType: OneToOne,
			participants: []*Participant{
				{ID: "1", ConnectionInfo: &ConnectionInfo{Type: OneToOne}},
				{ID: "2", ConnectionInfo: &ConnectionInfo{Type: OneToOne}},
			},
			expectError: false,
		},
		{
			name:     "OneToOne_Full",
			roomType: OneToOne,
			participants: []*Participant{
				{ID: "1", ConnectionInfo: &ConnectionInfo{Type: OneToOne}},
				{ID: "2", ConnectionInfo: &ConnectionInfo{Type: OneToOne}},
				{ID: "3", ConnectionInfo: &ConnectionInfo{Type: OneToOne}},
			},
			expectError:   true,
			errorExpected: ErrRoomFull,
		},
		{
			name:     "Broadcasting_Success",
			roomType: Broadcasting,
			participants: []*Participant{
				{ID: "1", ConnectionInfo: &ConnectionInfo{Type: Broadcasting, IsBroadcaster: true}},
				{ID: "2", ConnectionInfo: &ConnectionInfo{Type: Broadcasting}},
				{ID: "3", ConnectionInfo: &ConnectionInfo{Type: Broadcasting}},
			},
			expectError: false,
		},
		{
			name:     "Broadcasting_MultipleBroadcasters",
			roomType: Broadcasting,
			participants: []*Participant{
				{ID: "1", ConnectionInfo: &ConnectionInfo{Type: Broadcasting, IsBroadcaster: true}},
				{ID: "2", ConnectionInfo: &ConnectionInfo{Type: Broadcasting, IsBroadcaster: true}},
			},
			expectError:   true,
			errorExpected: ErrBroadcasterExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room := NewRoom("test-room", tt.roomType)

			var lastError error
			for _, p := range tt.participants {
				lastError = room.AddParticipant(p)
				if lastError != nil {
					break
				}
			}

			if tt.expectError {
				if lastError != tt.errorExpected {
					t.Errorf("Expected error %v, got %v", tt.errorExpected, lastError)
				}
			} else {
				if lastError != nil {
					t.Errorf("Expected no error, got %v", lastError)
				}
			}
		})
	}
}

func TestRemoveParticipant(t *testing.T) {
	room := NewRoom("test-room", Broadcasting)
	broadcaster := &Participant{
		ID: "broadcaster",
		ConnectionInfo: &ConnectionInfo{
			Type:          Broadcasting,
			IsBroadcaster: true,
		},
	}
	viewer := &Participant{
		ID: "viewer",
		ConnectionInfo: &ConnectionInfo{
			Type: Broadcasting,
		},
	}

	room.AddParticipant(broadcaster)
	room.AddParticipant(viewer)

	if room.Broadcaster == nil {
		t.Error("Expected broadcaster to be set")
	}

	room.RemoveParticipant(broadcaster.ID)

	if room.Broadcaster != nil {
		t.Error("Expected broadcaster to be removed")
	}

	if _, exists := room.Participants[broadcaster.ID]; exists {
		t.Error("Expected broadcaster to be removed from participants")
	}
}

func TestChatMessages(t *testing.T) {
	room := NewRoom("test-room", OneToOne)
	message := ChatMessage{
		SenderID:   "user1",
		SenderName: "User 1",
		Content:    "Hello, World!",
		Timestamp:  123456789,
	}

	room.AddChatMessage(message)

	history := room.GetChatHistory()
	if len(history) != 1 {
		t.Errorf("Expected 1 message in history, got %d", len(history))
	}

	if history[0].Content != message.Content {
		t.Errorf("Expected message content %s, got %s", message.Content, history[0].Content)
	}
}
