package models

// ConnectionType defines the type of connection in a room
type ConnectionType string

const (
	// OneToOne represents a direct peer-to-peer connection
	OneToOne ConnectionType = "one_to_one"
	// Broadcasting represents a one-to-many connection
	Broadcasting ConnectionType = "broadcasting"
	// ScreenSharing represents a screen sharing connection
	ScreenSharing ConnectionType = "screen_sharing"
)

// ConnectionInfo stores information about the connection
type ConnectionInfo struct {
	Type          ConnectionType `json:"type"`
	IsBroadcaster bool           `json:"is_broadcaster"`
	IsScreenShare bool           `json:"is_screen_share"`
}
