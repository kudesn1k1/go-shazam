<template>
  <div class="page">
    <header class="top-bar">
      <div class="branding">
        <span class="pulse-dot" aria-hidden="true" />
        <div>
          <strong>Go Shazam</strong>
          <p>Find the music around you</p>
        </div>
      </div>
      <div class="actions">
        <template v-if="isAuthenticated">
          <span class="user-email">{{ user?.email }}</span>
          <button class="ghost" @click="handleLogout">Log out</button>
        </template>
        <template v-else>
          <button class="ghost" @click="openAuthModal('login')">Log in</button>
          <button class="pill" @click="openAuthModal('register')">Create account</button>
        </template>
      </div>
    </header>

    <main class="hero">
      <section class="status-card">
        <p class="status-label">{{ statusMessage }}</p>

        <div v-if="detectedTrack" class="result-card">
          <p class="result-caption">We think it's</p>
          <h2>{{ detectedTrack.title }}</h2>
          <p class="artist">{{ detectedTrack.artist }}</p>
          <p v-if="detectedTrack.confidence" class="confidence">
            Confidence: {{ detectedTrack.confidence }}
          </p>

          <iframe v-if="detectedTrack.youtubeId" :src="`https://www.youtube.com/embed/${detectedTrack.youtubeId}?start=${Math.round(detectedTrack.timeOffset)}`" width="100%" height="100%" frameborder="0" allowfullscreen></iframe>
        </div>

        <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
      </section>

      <button
        class="record-btn"
        :class="{ 'is-recording': isRecording, 'is-loading': isSending }"
        @click="toggleRecording"
        :disabled="isSending"
        aria-live="polite"
      >
        <span>{{ recordButtonLabel }}</span>
        <small v-if="isRecording">Tap to stop</small>
      </button>

      <p class="hint">Tip: hold your device close to the speaker for the best results.</p>
    </main>

    <AuthModal
      v-model="isAuthModalOpen"
      :mode="authMode"
      :pending="authLoading"
      @submit="handleAuthSubmit"
    />

    <transition name="toast">
      <div v-if="authToast" class="toast" :class="{ 'toast-error': authToastType === 'error' }">
        {{ authToast }}
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';

import AuthModal from './components/AuthModal.vue';
import { useAuth } from './composables/useAuth';

type RecognitionResult = {
  title: string;
  artist: string;
  timeOffset?: number;
  confidence?: number;
  youtubeId?: string;
};

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:5000';
const MAX_RECORDING_MS = 12_000;

const isRecording = ref(false);
const isSending = ref(false);
const statusMessage = ref('Tap to listen for music around you');
const errorMessage = ref('');
const detectedTrack = ref<RecognitionResult | null>(null);

const isAuthModalOpen = ref(false);
const authMode = ref<'login' | 'register'>('login');
const authToast = ref('');
const authToastType = ref<'success' | 'error'>('success');

const {
  user,
  isAuthenticated,
  isLoading: authLoading,
  error: authError,
  login,
  register,
  logout,
  initialize: initAuth,
} = useAuth();

let audioContext: AudioContext | null = null;
let processor: ScriptProcessorNode | null = null;
let socket: WebSocket | null = null;
let stream: MediaStream | null = null;
let autoStopTimer: ReturnType<typeof setTimeout> | null = null;

// Initialize auth on mount
onMounted(async () => {
  await initAuth();
});

const recordButtonLabel = computed(() => {
  if (isRecording.value) {
    return 'Listening…';
  }
  if (isSending.value) {
    return 'Identifying…';
  }

  return 'Tap to Shazam';
});

const resetRecorder = () => {
  if (processor) {
    processor.disconnect();
    processor = null;
  }
  if (audioContext) {
    audioContext.close();
    audioContext = null;
  }
  if (stream) {
    stream.getTracks().forEach((track) => track.stop());
    stream = null;
  }
  if (socket) {
    socket.close();
    socket = null;
  }

  if (autoStopTimer) {
    clearTimeout(autoStopTimer);
    autoStopTimer = null;
  }
};

const stopRecording = () => {
  if (!isRecording.value) {
    return;
  }

  isRecording.value = false;
  statusMessage.value = 'Processing audio…';
  isSending.value = true;

  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send('stop');
  }
  
  // Don't close socket immediately, wait for response
  if (processor) {
    processor.disconnect();
    processor = null;
  }
  if (audioContext) {
    audioContext.close();
    audioContext = null;
  }
  if (stream) {
    stream.getTracks().forEach((track) => track.stop());
    stream = null;
  }
};

const startRecording = async () => {
  if (isRecording.value || isSending.value) {
    return;
  }

  if (!navigator.mediaDevices?.getUserMedia) {
    errorMessage.value = 'Microphone access is not supported in this browser.';
    return;
  }

  try {
    stream = await navigator.mediaDevices.getUserMedia({
      audio: {
        channelCount: 1,
        sampleRate: 44100,
        echoCancellation: true,
        noiseSuppression: true,
      },
    });

    audioContext = new (window.AudioContext || (window as any).webkitAudioContext)();
    const source = audioContext.createMediaStreamSource(stream);
    processor = audioContext.createScriptProcessor(4096, 1, 1);

    source.connect(processor);
    processor.connect(audioContext.destination);

    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    // Use specific port if needed or derive from API_BASE_URL
    const wsBase = API_BASE_URL.replace(/^http/, 'ws');
    socket = new WebSocket(`${wsBase}/api/recognize/ws`);
    socket.binaryType = 'arraybuffer';

    socket.onopen = () => {
      if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(`start:${audioContext?.sampleRate || 44100}`);
        isRecording.value = true;
        statusMessage.value = 'Listening… hold your phone close to the source';
        errorMessage.value = '';
        detectedTrack.value = null;

        autoStopTimer = setTimeout(() => {
          stopRecording();
        }, MAX_RECORDING_MS);
      }
    };

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.error) {
          errorMessage.value = data.error;
          statusMessage.value = 'Error occurred';
        } else if (data.found && data.song) {
          detectedTrack.value = {
            title: data.song.Title,
            artist: data.song.Artist,
            youtubeId: data.song.SourceID,
            timeOffset: data.time_offset,
            confidence: data.score,
          };
          statusMessage.value = 'Here is what we found';
        } else {
          statusMessage.value = 'Could not identify this track';
        }
      } catch (e) {
        console.error(e);
      } finally {
        isSending.value = false;
        resetRecorder();
      }
    };

    socket.onerror = (e) => {
        console.error("WebSocket error", e);
        errorMessage.value = "Connection error";
        isRecording.value = false;
        isSending.value = false;
        resetRecorder();
    };

    processor.onaudioprocess = (e) => {
      if (!isRecording.value) return;
      const inputData = e.inputBuffer.getChannelData(0);
      if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(inputData);
      }
    };

  } catch (error) {
    errorMessage.value =
      error instanceof DOMException ? error.message : 'Unable to access microphone.';
    resetRecorder();
  }
};

const toggleRecording = () => {
  if (isRecording.value) {
    stopRecording();
  } else {
    void startRecording();
  }
};

const openAuthModal = (mode: 'login' | 'register') => {
  authMode.value = mode;
  isAuthModalOpen.value = true;
};

const showToast = (message: string, type: 'success' | 'error' = 'success') => {
  authToast.value = message;
  authToastType.value = type;
  setTimeout(() => {
    authToast.value = '';
  }, 4000);
};

const handleAuthSubmit = async (payload: { mode: 'login' | 'register'; email: string; password: string }) => {
  const success = payload.mode === 'login'
    ? await login(payload.email, payload.password)
    : await register(payload.email, payload.password);

  if (success) {
    isAuthModalOpen.value = false;
    showToast(
      payload.mode === 'login' ? 'Welcome back!' : 'Account created successfully!',
      'success'
    );
  } else {
    showToast(authError.value || 'An error occurred', 'error');
  }
};

const handleLogout = async () => {
  await logout();
  showToast('You have been logged out', 'success');
};

onBeforeUnmount(() => {
  resetRecorder();
});
</script>

<style scoped>
.user-email {
  font-size: 0.9rem;
  opacity: 0.8;
  margin-right: 0.5rem;

  display: flex;
  align-items: center;
}

.toast-error {
  background: rgba(255, 78, 78, 0.15) !important;
  border-color: rgba(255, 78, 78, 0.3) !important;
  color: #ff9f9f !important;
}
</style>
