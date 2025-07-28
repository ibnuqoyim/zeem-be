package services

import (
	"fmt"
	"log"
	"sync"

	"zeem/internal/models"

	"github.com/pion/webrtc/v3"
)

type SFUManager struct {
	mu              sync.RWMutex
	participants    map[string]*models.Participant
	peerConnections map[string]*webrtc.PeerConnection
	trackLocals     map[string]map[string]*webrtc.TrackLocalStaticRTP // participantID -> trackID -> track
}

func NewSFUManager() *SFUManager {
	return &SFUManager{
		participants:    make(map[string]*models.Participant),
		peerConnections: make(map[string]*webrtc.PeerConnection),
		trackLocals:     make(map[string]map[string]*webrtc.TrackLocalStaticRTP),
	}
}

func (s *SFUManager) AddParticipant(participant *models.Participant) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a new PeerConnection
	me := webrtc.MediaEngine{}
	if err := me.RegisterDefaultCodecs(); err != nil {
		return err
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(&me))
	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		return err
	}

	s.participants[participant.ID] = participant
	s.peerConnections[participant.ID] = peerConnection
	s.trackLocals[participant.ID] = make(map[string]*webrtc.TrackLocalStaticRTP)

	// Handle ICE connection state
	peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Printf("ICE Connection State has changed: %s", state.String())
	})

	// Handle tracks
	peerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		log.Printf("Track received from participant %s", participant.ID)
		s.handleTrack(participant.ID, remoteTrack)
	})

	log.Printf("Participant added: %s", participant.Username)
	return nil
}

func (s *SFUManager) RemoveParticipant(participantID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if participant, exists := s.participants[participantID]; exists {
		// Close and remove peer connection
		if pc, ok := s.peerConnections[participantID]; ok {
			pc.Close()
			delete(s.peerConnections, participantID)
		}

		// Remove all tracks associated with this participant
		delete(s.trackLocals, participantID)

		// Remove participant
		delete(s.participants, participantID)
		log.Printf("Participant removed: %s", participant.Username)
	}
}

func (s *SFUManager) handleTrack(senderID string, remoteTrack *webrtc.TrackRemote) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a new local track
	trackLocal, err := webrtc.NewTrackLocalStaticRTP(remoteTrack.Codec().RTPCodecCapability, remoteTrack.ID(), remoteTrack.StreamID())
	if err != nil {
		log.Printf("Failed to create new track: %v", err)
		return
	}

	// Store the track
	s.trackLocals[senderID][remoteTrack.ID()] = trackLocal

	// Forward the track to all other participants
	for participantID, pc := range s.peerConnections {
		if participantID == senderID {
			continue
		}

		// Add the track to the peer connection
		if _, err = pc.AddTrack(trackLocal); err != nil {
			log.Printf("Failed to add track to peer %s: %v", participantID, err)
			continue
		}

		log.Printf("Track forwarded to participant %s", participantID)
	}

	// Start forwarding RTP packets
	go func() {
		for {
			packet, _, err := remoteTrack.ReadRTP()
			if err != nil {
				return
			}

			// Write RTP packets to all track locals
			for participantID := range s.peerConnections {
				if participantID == senderID {
					continue
				}

				if track, ok := s.trackLocals[senderID][remoteTrack.ID()]; ok {
					if err := track.WriteRTP(packet); err != nil {
						log.Printf("Failed to write RTP to participant %s: %v", participantID, err)
					}
				}
			}
		}
	}()
}

func (s *SFUManager) HandleOffer(participantID string, offer webrtc.SessionDescription) (webrtc.SessionDescription, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pc, exists := s.peerConnections[participantID]
	if !exists {
		return webrtc.SessionDescription{}, fmt.Errorf("no peer connection found for participant %s", participantID)
	}

	if err := pc.SetRemoteDescription(offer); err != nil {
		return webrtc.SessionDescription{}, err
	}

	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	if err = pc.SetLocalDescription(answer); err != nil {
		return webrtc.SessionDescription{}, err
	}

	return answer, nil
}
