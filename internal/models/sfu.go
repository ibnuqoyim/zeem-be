package models

// Stream represents a media stream in the SFU.
type Stream struct {
	ID           string   `json:"id"`
	Participants []string `json:"participants"`
}

// SFUConfig holds the configuration for the SFU.
type SFUConfig struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}
