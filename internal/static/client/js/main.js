let webrtcClient;

// Add event listeners when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    document.getElementById('joinButton').addEventListener('click', joinRoom);
});

async function joinRoom() {
    try {
        const username = document.getElementById('username').value;
        const roomId = document.getElementById('roomId').value;

        if (!username || !roomId) {
            alert('Please enter both username and room ID');
            return;
        }

        // Check browser compatibility
        if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
            alert('Your browser does not support WebRTC. Please use a modern browser like Chrome, Firefox, or Safari.');
            return;
        }

        webrtcClient = new WebRTCClient();
        
        // Request permissions before initializing
        try {
            const mediaStream = await navigator.mediaDevices.getUserMedia({
                audio: true,
                video: { facingMode: "user" } // "user" = kamera depan, "environment" = kamera belakang
            });
        } catch (error) {
            if (error.name === 'NotAllowedError' || error.name === 'PermissionDeniedError') {
                alert('Please allow access to your camera and microphone to join the room.');
            } else {
                alert('Error accessing media devices: ' + error.message);
            }
            return;
        }

        const success = await webrtcClient.initialize(username, roomId);

        if (success) {
            document.getElementById('joinForm').classList.add('hidden');
            document.getElementById('meetingRoom').classList.remove('hidden');
            setupControls();
        } else {
            alert('Failed to join room. Please try again.');
        }
    } catch (error) {
        console.error('Error joining room:', error);
        alert('Error joining room: ' + error.message);
    }
}

function setupControls() {
    document.getElementById('toggleVideo').addEventListener('click', async () => {
        const enabled = await webrtcClient.toggleVideo();
        document.getElementById('toggleVideo').textContent = 
            enabled ? 'Turn Off Video' : 'Turn On Video';
    });

    document.getElementById('toggleAudio').addEventListener('click', async () => {
        const enabled = await webrtcClient.toggleAudio();
        document.getElementById('toggleAudio').textContent = 
            enabled ? 'Mute Audio' : 'Unmute Audio';
    });

    document.getElementById('shareScreen').addEventListener('click', async () => {
        await webrtcClient.shareScreen();
    });

    document.getElementById('leaveRoom').addEventListener('click', leaveRoom);
}

// Handle page unload
window.addEventListener('beforeunload', () => {
    if (webrtcClient) {
        webrtcClient.disconnect();
    }
});

function leaveRoom() {
    if (webrtcClient) {
        webrtcClient.disconnect();
        document.getElementById('videoGrid').innerHTML = '';
        document.getElementById('meetingRoom').classList.add('hidden');
        document.getElementById('joinForm').classList.remove('hidden');
    }
}

// Handle page unload
window.addEventListener('beforeunload', () => {
    if (webrtcClient) {
        webrtcClient.disconnect();
    }
});
