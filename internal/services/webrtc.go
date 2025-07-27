package services

import (
	"log"
	"sync"

	"github.com/pion/webrtc/v3"
)

// WebRTCManager handles WebRTC peer connections
type WebRTCManager struct {
	peerConnections map[string]*webrtc.PeerConnection
	mutex           sync.RWMutex
	config          webrtc.Configuration
}

// NewWebRTCManager creates a new WebRTC manager
func NewWebRTCManager() *WebRTCManager {
	// Configure ICE servers (STUN/TURN)
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	return &WebRTCManager{
		peerConnections: make(map[string]*webrtc.PeerConnection),
		config:          config,
	}
}

// CreatePeerConnection creates a new WebRTC peer connection
func (m *WebRTCManager) CreatePeerConnection(peerID string) (*webrtc.PeerConnection, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Create a new peer connection
	peerConnection, err := webrtc.NewPeerConnection(m.config)
	if err != nil {
		return nil, err
	}

	// Set the handler for ICE connection state
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Printf("ICE Connection State has changed: %s\n", connectionState.String())
	})

	// Set the handler for connection state
	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		log.Printf("Peer Connection State has changed: %s\n", s.String())
		if s == webrtc.PeerConnectionStateFailed {
			if err := peerConnection.Close(); err != nil {
				log.Printf("Error closing peer connection: %v\n", err)
			}
			m.RemovePeerConnection(peerID)
		}
	})

	// Store the peer connection
	m.peerConnections[peerID] = peerConnection
	return peerConnection, nil
}

// GetPeerConnection retrieves an existing peer connection
func (m *WebRTCManager) GetPeerConnection(peerID string) *webrtc.PeerConnection {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.peerConnections[peerID]
}

// RemovePeerConnection removes a peer connection
func (m *WebRTCManager) RemovePeerConnection(peerID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if pc := m.peerConnections[peerID]; pc != nil {
		if err := pc.Close(); err != nil {
			log.Printf("Error closing peer connection: %v\n", err)
		}
	}
	delete(m.peerConnections, peerID)
}

// HandleOffer processes a WebRTC offer
func (m *WebRTCManager) HandleOffer(peerID string, offerSDP string) (*webrtc.SessionDescription, error) {
	pc, err := m.CreatePeerConnection(peerID)
	if err != nil {
		return nil, err
	}

	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offerSDP,
	}

	if err := pc.SetRemoteDescription(offer); err != nil {
		return nil, err
	}

	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	if err := pc.SetLocalDescription(answer); err != nil {
		return nil, err
	}

	return &answer, nil
}

// HandleAnswer processes a WebRTC answer
func (m *WebRTCManager) HandleAnswer(peerID string, answerSDP string) error {
	pc := m.GetPeerConnection(peerID)
	if pc == nil {
		return nil
	}

	answer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  answerSDP,
	}

	return pc.SetRemoteDescription(answer)
}

// HandleICECandidate processes a WebRTC ICE candidate
func (m *WebRTCManager) HandleICECandidate(peerID string, candidate string) error {
	pc := m.GetPeerConnection(peerID)
	if pc == nil {
		return nil
	}

	iceCandidate := webrtc.ICECandidateInit{Candidate: candidate}
	return pc.AddICECandidate(iceCandidate)
}
