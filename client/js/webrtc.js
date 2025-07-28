class WebRTCClient {
    constructor() {
        this.socket = null;
        this.peerConnection = null;
        this.localStream = null;
        this.remoteStreams = new Map();
        this.username = '';
        this.roomId = '';
        
        this.mediaConstraints = {
            audio: true,
            video: {
                width: { ideal: 1280 },
                height: { ideal: 720 }
            }
        };
    }

    async initialize(username, roomId) {
        this.username = username;
        this.roomId = roomId;
        
        try {
            // Check if getUserMedia is supported
            if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
                throw new Error('Your browser does not support media devices. Please use a modern browser.');
            }

            // Request permissions first
            await navigator.mediaDevices.getUserMedia({ audio: true, video: true });
            
            // Get local media stream with desired constraints
            this.localStream = await navigator.mediaDevices.getUserMedia(this.mediaConstraints);
            this.addVideoStream('local', this.localStream, username);

            // Create WebSocket connection
            await this.connectSignalingServer();

            // Create and setup RTCPeerConnection
            await this.createPeerConnection();

            // Add local tracks to peer connection
            await this.addTracksToPeerConnection();
            console.log('Local tracks added to peer connection');

            return true;
        } catch (error) {
            console.error('Error initializing WebRTC:', error);
            return false;
        }
    }

    async connectSignalingServer() {
        return new Promise((resolve, reject) => {
            // Always use secure WebSocket for HTTPS
            const wsUrl = `wss://${window.location.hostname}:8443/ws?roomId=${this.roomId}&username=${this.username}`;
            
            this.socket = new WebSocket(wsUrl);

            this.socket.onopen = () => {
                console.log('Connected to signaling server');
                this.setupSignalingHandlers();
                resolve();
            };

            this.socket.onerror = (error) => {
                console.error('WebSocket error:', error);
                reject(error);
            };
        });
    }

    async createPeerConnection() {
        const configuration = {
            iceServers: [
                { urls: 'stun:stun.l.google.com:19302' },
                { urls: 'stun:stun1.l.google.com:19302' },
                { urls: 'stun:stun2.l.google.com:19302' },
                { urls: 'stun:stun3.l.google.com:19302' },
                { urls: 'stun:stun4.l.google.com:19302' }
            ]
        };

        this.peerConnection = new RTCPeerConnection(configuration);

        // Handle ICE candidates
        this.peerConnection.onicecandidate = (event) => {
            if (event.candidate) {
                this.socket.send(JSON.stringify({
                    type: 'ice_candidate',
                    data: event.candidate,
                    roomId: this.roomId
                }));
            }
        };

        // Handle connection state changes
        this.peerConnection.onconnectionstatechange = () => {
            console.log('Connection state:', this.peerConnection.connectionState);
        };

        // Handle ICE connection state changes
        this.peerConnection.oniceconnectionstatechange = () => {
            console.log('ICE Connection state:', this.peerConnection.iceConnectionState);
        };

        // Handle negotiation needed
        this.peerConnection.onnegotiationneeded = async () => {
            try {
                const offer = await this.peerConnection.createOffer();
                await this.peerConnection.setLocalDescription(offer);
                this.socket.send(JSON.stringify({
                    type: 'offer',
                    data: offer,
                    roomId: this.roomId
                }));
            } catch (error) {
                console.error('Error creating offer:', error);
            }
        };

        // Handle incoming tracks
        this.peerConnection.ontrack = (event) => {
            console.log('Received remote track:', event.streams[0].id);
            const remoteStream = event.streams[0];
            if (!this.remoteStreams.has(remoteStream.id)) {
                console.log('Adding new remote stream');
                this.remoteStreams.set(remoteStream.id, remoteStream);
                
                // Create a new video element for this stream
                const videoId = `remote-${remoteStream.id}`;
                this.addVideoStream(videoId, remoteStream, 'Remote User');
                
                // Log track information
                event.streams[0].getTracks().forEach(track => {
                    console.log('Remote track details:', {
                        kind: track.kind,
                        enabled: track.enabled,
                        id: track.id
                    });
                });
            }
        };
    }

    setupSignalingHandlers() {
        this.socket.onmessage = async (event) => {
            const message = JSON.parse(event.data);
            console.log('Received message:', message.type);

            switch (message.type) {
                case 'offer':
                    console.log('Received offer from remote peer');
                    await this.handleOffer(message);
                    break;
                case 'answer':
                    console.log('Received answer from remote peer');
                    await this.handleAnswer(message);
                    break;
                case 'ice_candidate':
                    console.log('Received ICE candidate');
                    await this.handleIceCandidate(message);
                    break;
                case 'participant_joined':
                    console.log('New participant joined:', message.data.username);
                    // Create offer for new participant
                    try {
                        const offer = await this.peerConnection.createOffer({
                            offerToReceiveAudio: true,
                            offerToReceiveVideo: true
                        });
                        await this.peerConnection.setLocalDescription(offer);
                        this.socket.send(JSON.stringify({
                            type: 'offer',
                            data: offer,
                            roomId: this.roomId
                        }));
                    } catch (error) {
                        console.error('Error creating offer for new participant:', error);
                    }
                    break;
                case 'participant_left':
                    this.handleParticipantLeft(message);
                    break;
            }
        };
    }

    async handleOffer(message) {
        try {
            if (this.peerConnection.signalingState !== "stable") {
                console.log("Signaling state is not stable, waiting...");
                return;
            }

            await this.peerConnection.setRemoteDescription(new RTCSessionDescription(message.data));
            
            // Make sure we have local stream
            if (!this.localStream) {
                this.localStream = await navigator.mediaDevices.getUserMedia(this.mediaConstraints);
                this.addVideoStream('local', this.localStream, this.username);
                await this.addTracksToPeerConnection();
            }

            const answer = await this.peerConnection.createAnswer();
            await this.peerConnection.setLocalDescription(answer);

            this.socket.send(JSON.stringify({
                type: 'answer',
                data: answer,
                roomId: this.roomId
            }));
        } catch (error) {
            console.error('Error handling offer:', error);
        }
    }

    async handleAnswer(message) {
        try {
            console.log('Setting remote description (answer)');
            await this.peerConnection.setRemoteDescription(new RTCSessionDescription(message.data));
            console.log('Remote description set successfully');
        } catch (error) {
            console.error('Error handling answer:', error);
        }
    }

    async handleIceCandidate(message) {
        try {
            console.log('Adding ICE candidate');
            await this.peerConnection.addIceCandidate(new RTCIceCandidate(message.data));
            console.log('ICE candidate added successfully');
        } catch (error) {
            console.error('Error handling ICE candidate:', error);
        }
    }

    async addTracksToPeerConnection() {
        if (this.localStream) {
            this.localStream.getTracks().forEach(track => {
                console.log('Adding track to peer connection:', track.kind);
                this.peerConnection.addTrack(track, this.localStream);
            });
        }
    }

    handleParticipantLeft(message) {
        const videoElement = document.getElementById(`video-${message.senderId}`);
        if (videoElement) {
            videoElement.parentElement.remove();
        }
    }

    addVideoStream(id, stream, username) {
        const videoGrid = document.getElementById('videoGrid');
        const videoContainer = document.createElement('div');
        videoContainer.className = 'video-container';
        videoContainer.id = `container-${id}`;

        const video = document.createElement('video');
        video.id = `video-${id}`;
        video.srcObject = stream;
        video.autoplay = true;
        video.playsInline = true;
        if (id === 'local') {
            video.muted = true;
        }

        const usernameLabel = document.createElement('div');
        usernameLabel.className = 'username-label';
        usernameLabel.textContent = username;

        videoContainer.appendChild(video);
        videoContainer.appendChild(usernameLabel);
        videoGrid.appendChild(videoContainer);
    }

    async toggleVideo() {
        const videoTrack = this.localStream.getVideoTracks()[0];
        if (videoTrack) {
            videoTrack.enabled = !videoTrack.enabled;
            return videoTrack.enabled;
        }
        return false;
    }

    async toggleAudio() {
        const audioTrack = this.localStream.getAudioTracks()[0];
        if (audioTrack) {
            audioTrack.enabled = !audioTrack.enabled;
            return audioTrack.enabled;
        }
        return false;
    }

    async shareScreen() {
        try {
            const screenStream = await navigator.mediaDevices.getDisplayMedia({ video: true });
            const videoTrack = screenStream.getVideoTracks()[0];

            // Replace video track
            const sender = this.peerConnection.getSenders().find(s => s.track.kind === 'video');
            await sender.replaceTrack(videoTrack);

            // Update local video
            const localVideo = document.getElementById('video-local');
            localVideo.srcObject = screenStream;

            // Handle stop sharing
            videoTrack.onended = async () => {
                const cameraTrack = this.localStream.getVideoTracks()[0];
                await sender.replaceTrack(cameraTrack);
                localVideo.srcObject = this.localStream;
            };

            return true;
        } catch (error) {
            console.error('Error sharing screen:', error);
            return false;
        }
    }

    disconnect() {
        if (this.localStream) {
            this.localStream.getTracks().forEach(track => track.stop());
        }
        if (this.peerConnection) {
            this.peerConnection.close();
        }
        if (this.socket) {
            this.socket.close();
        }
    }
}
