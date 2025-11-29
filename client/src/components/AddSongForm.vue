<template>
  <div class="add-song-wrapper">
    <div class="add-song-input-group">
      <input 
        v-model="addSongLink" 
        type="text" 
        placeholder="Paste Spotify link to add song..." 
        class="song-input"
        @keyup.enter="handleAddSong"
      />
      <button class="pill small" :disabled="isAddingSong" @click="handleAddSong">
        {{ isAddingSong ? 'Adding...' : 'Add Song' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useAuth } from '../composables/useAuth';

const emit = defineEmits<{
  (e: 'require-auth'): void;
  (e: 'toast', message: string, type: 'success' | 'error'): void;
}>();

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:5000';
const addSongLink = ref('');
const isAddingSong = ref(false);

const { isAuthenticated, getAccessToken } = useAuth();

const handleAddSong = async () => {
  if (!isAuthenticated.value) {
    emit('require-auth');
    return;
  }
  
  if (!addSongLink.value) {
    emit('toast', 'Please enter a link', 'error');
    return;
  }

  // Basic Spotify link validation
  if (!addSongLink.value.includes('spotify.com')) {
    emit('toast', 'Please enter a valid Spotify link', 'error');
    return;
  }

  isAddingSong.value = true;
  
  try {
    const response = await fetch(`${API_BASE_URL}/api/song/add`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${getAccessToken()}` 
        },
        body: JSON.stringify({ link: addSongLink.value })
    });

    const data = await response.json();

    if (response.ok) {
        addSongLink.value = '';
        emit('toast', 'Song added to queue! Processing started.', 'success');
    } else {
        emit('toast', data.error || 'Failed to add song', 'error');
    }

  } catch (e) {
      emit('toast', 'Network error', 'error');
  } finally {
      isAddingSong.value = false;
  }
};
</script>

<style scoped>
.add-song-wrapper {
  margin-top: 2rem;
  width: 100%;
  max-width: 400px;
  display: flex;
  justify-content: center;
  padding: 0 1rem;
}

.add-song-input-group {
  display: flex;
  gap: 0.5rem;
  width: 100%;
}

.song-input {
  flex: 1;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 999px;
  padding: 0.65rem 1.2rem;
  color: #fff;
  font-family: inherit;
  font-size: 0.95rem;
  outline: none;
  transition: border-color 0.2s, background 0.2s;
  min-width: 0; /* Prevent flex overflow */
}

.song-input:focus {
  border-color: rgba(0, 209, 255, 0.6);
  background: rgba(255, 255, 255, 0.15);
}

.song-input::placeholder {
  color: rgba(255, 255, 255, 0.5);
}

.pill.small {
  padding: 0.65rem 1.2rem;
  font-size: 0.9rem;
  white-space: nowrap;
}
</style>

