# Zeem Video Conference - Technical Documentation

## Architecture Overview

### Backend Components
1. **Server (Go/Gin)**
   - WebSocket Signaling Server
   - HTTPS Server
   - Static File Server
   - Security Middleware
   - Room Management
   - WebRTC Management

2. **WebRTC SFU (Selective Forwarding Unit)**
   - Media Stream Management
   - Peer Connection Management
   - ICE Candidate Handling
   - Track Management

### Frontend Components
1. **Web Client**
   - WebRTC Client
   - Media Stream Handling
   - User Interface
   - Real-time Communication

## Technical Specifications

### Backend (Go)

#### Server Requirements
- Go 1.21 or higher
- SSL Certificate (for HTTPS)
- Open ports: 8443 (HTTPS/WSS)

#### Dependencies
```go
require (
    github.com/gin-gonic/gin
    github.com/gorilla/websocket
    github.com/pion/webrtc/v3
)
```

#### Key Features
- Secure WebSocket Communication
- Room-based Video Conference
- Multiple Participant Support
- Screen Sharing
- Audio/Video Controls

### Frontend (JavaScript)

#### Browser Requirements
- Modern browsers (Chrome, Firefox, Safari)
- WebRTC support
- Camera and Microphone access

#### Key Classes
1. **WebRTCClient**
   ```javascript
   class WebRTCClient {
       constructor()
       initialize(username, roomId)
       connectSignalingServer()
       createPeerConnection()
       handleOffer(message)
       handleAnswer(message)
       handleIceCandidate(message)
   }
   ```

## Security Implementation

### Server-side Security
1. **HTTPS/WSS**
   - Secure communication
   - SSL/TLS encryption

2. **Content Security Policy (CSP)**
   ```http
   default-src 'self';
   script-src 'self' 'unsafe-inline' 'unsafe-eval';
   style-src 'self' 'unsafe-inline';
   connect-src 'self' wss: https: ws:;
   media-src 'self' blob: mediastream:;
   ```

3. **CORS Policy**
   - Strict origin checking
   - Credential support

### WebRTC Security
1. **STUN/TURN Configuration**
   - Multiple STUN servers
   - ICE candidate verification

2. **Media Security**
   - Encrypted media streams
   - Secure key exchange

## API Documentation

### WebSocket Messages

1. **Join Room**
   ```json
   {
     "type": "participant_joined",
     "data": {
       "username": "string",
       "roomId": "string"
     }
   }
   ```

2. **WebRTC Signaling**
   ```json
   {
     "type": "offer|answer|ice_candidate",
     "data": {},
     "roomId": "string"
   }
   ```

## Directory Structure
```
zeem-be/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── handlers/
│   │   ├── websocket.go
│   │   └── sfu.go
│   ├── models/
│   │   └── room.go
│   └── services/
│       ├── room_manager.go
│       ├── webrtc.go
│       └── sfu_manager.go
├── client/
│   ├── index.html
│   ├── css/
│   │   └── style.css
│   └── js/
│       ├── main.js
│       └── webrtc.js
├── certs/
│   ├── cert.pem
│   └── key.pem
└── docs/
    └── technical_documentation.md
```
