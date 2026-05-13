<template>
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

        <iframe
          v-if="detectedTrack.youtubeId"
          :src="`https://www.youtube.com/embed/${detectedTrack.youtubeId}?start=${Math.round(detectedTrack.timeOffset ?? 0)}`"
          width="100%"
          height="100%"
          frameborder="0"
          allowfullscreen
        />
      </div>

      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
    </section>

    <div class="record-btn-wrapper">
      <div
        v-if="isRecording"
        class="audio-waves"
        :style="{ '--audio-level': audioLevel }"
      >
        <div class="wave wave-1" />
        <div class="wave wave-2" />
        <div class="wave wave-3" />
      </div>
      <button
        class="record-btn"
        :class="{ 'is-recording': isRecording, 'is-loading': isSending }"
        :style="isRecording ? {
          transform: `scale(${1 + audioLevel * 0.15})`,
          boxShadow: `0 20px ${80 + audioLevel * 40}px rgba(255, 78, 126, ${0.6 + audioLevel * 0.4})`
        } : {}"
        :disabled="isSending"
        aria-live="polite"
        @click="toggleRecording"
      >
        <span>{{ recordButtonLabel }}</span>
        <small v-if="isRecording">Tap to stop</small>
      </button>
    </div>

    <AddSongForm @require-auth="showAuthPrompt" @toast="showToast" />

    <p class="hint">Tip: hold your device close to the speaker for the best results.</p>
  </main>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue';
import { useHead } from '@unhead/vue';
import AddSongForm from '../components/AddSongForm.vue';
import { useAuthModal } from '../composables/useAuthModal';
import { useToast } from '../composables/useToast';

useHead({
  title: 'Go Shazam — Identify the music around you',
  meta: [
    { name: 'description', content: 'Tap to identify any song playing around you. Browse our growing catalog of recognized tracks and explore by artist, title, or recency.' },
    { property: 'og:title', content: 'Go Shazam — Identify the music around you' },
    { property: 'og:description', content: 'Tap to identify any song playing around you.' },
    { property: 'og:type', content: 'website' },
  ],
  link: [
    { rel: 'canonical', href: () => (typeof window !== 'undefined' ? `${window.location.origin}/` : '/') },
  ],
});

const authModal = useAuthModal();
const toast = useToast();

type RecognitionResult = {
  title: string;
  artist: string;
  timeOffset?: number;
  confidence?: number;
  youtubeId?: string;
};

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:5000';
const MAX_RECORDING_MS = 8_000;

const isRecording = ref(false);
const isSending = ref(false);
const statusMessage = ref('Tap to listen for music around you');
const errorMessage = ref('');
const detectedTrack = ref<RecognitionResult | null>(null);
const audioLevel = ref(0);

let audioContext: AudioContext | null = null;
let processor: ScriptProcessorNode | null = null;
let socket: WebSocket | null = null;
let stream: MediaStream | null = null;
let autoStopTimer: ReturnType<typeof setTimeout> | null = null;

const recordButtonLabel = computed(() => {
  if (isRecording.value) return 'Listening…';
  if (isSending.value) return 'Identifying…';
  return 'Tap to Shazam';
});

const resetRecorder = () => {
  processor?.disconnect();
  processor = null;
  audioContext?.close();
  audioContext = null;
  stream?.getTracks().forEach((t) => t.stop());
  stream = null;
  socket?.close();
  socket = null;
  if (autoStopTimer) {
    clearTimeout(autoStopTimer);
    autoStopTimer = null;
  }
  audioLevel.value = 0;
};

const stopRecording = () => {
  if (!isRecording.value) return;
  isRecording.value = false;
  audioLevel.value = 0;
  statusMessage.value = 'Processing audio…';
  isSending.value = true;

  if (socket?.readyState === WebSocket.OPEN) socket.send('stop');

  processor?.disconnect();
  processor = null;
  audioContext?.close();
  audioContext = null;
  stream?.getTracks().forEach((t) => t.stop());
  stream = null;
};

const startRecording = async () => {
  if (isRecording.value || isSending.value) return;

  if (!navigator.mediaDevices?.getUserMedia) {
    errorMessage.value = 'Microphone access is not supported in this browser.';
    return;
  }

  try {
    stream = await navigator.mediaDevices.getUserMedia({
      audio: { channelCount: 1, sampleRate: 44100, echoCancellation: true, noiseSuppression: true },
    });

    audioContext = new (window.AudioContext || (window as any).webkitAudioContext)();
    const source = audioContext.createMediaStreamSource(stream);
    processor = audioContext.createScriptProcessor(4096, 1, 1);
    source.connect(processor);
    processor.connect(audioContext.destination);

    const wsBase = API_BASE_URL.replace(/^http/, 'ws');
    socket = new WebSocket(`${wsBase}/api/recognize/ws`);
    socket.binaryType = 'arraybuffer';

    socket.onopen = () => {
      if (socket?.readyState === WebSocket.OPEN) {
        socket.send(`start:${audioContext?.sampleRate || 44100}`);
        isRecording.value = true;
        statusMessage.value = 'Listening… hold your phone close to the source';
        errorMessage.value = '';
        detectedTrack.value = null;
        autoStopTimer = setTimeout(stopRecording, MAX_RECORDING_MS);
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
      console.error('WebSocket error', e);
      errorMessage.value = 'Connection error';
      isRecording.value = false;
      isSending.value = false;
      resetRecorder();
    };

    processor.onaudioprocess = (e) => {
      if (!isRecording.value) return;
      const inputData = e.inputBuffer.getChannelData(0);
      let sum = 0;
      for (let i = 0; i < inputData.length; i++) {
        const v = inputData[i] ?? 0;
        sum += v * v;
      }
      const rms = Math.sqrt(sum / inputData.length);
      audioLevel.value = audioLevel.value * 0.7 + Math.min(rms * 3, 1) * 0.3;
      if (socket?.readyState === WebSocket.OPEN) socket.send(inputData);
    };
  } catch (err) {
    errorMessage.value = err instanceof DOMException ? err.message : 'Unable to access microphone.';
    resetRecorder();
  }
};

const toggleRecording = () => {
  if (isRecording.value) stopRecording();
  else void startRecording();
};

const showAuthPrompt = () => {
  authModal.open('login');
};

const showToast = (message: string, type: 'success' | 'error' = 'success') => {
  toast.show(message, type);
};

onBeforeUnmount(resetRecorder);
</script>
