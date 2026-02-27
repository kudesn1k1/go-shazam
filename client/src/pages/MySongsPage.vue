<template>
  <div class="page-content">
    <h1 class="page-title">My Songs</h1>

    <div v-if="loading" class="state-msg">Loading…</div>
    <div v-else-if="error" class="state-msg error">{{ error }}</div>
    <div v-else-if="songs.length === 0" class="state-msg">You haven't uploaded any songs yet.</div>

    <template v-else>
      <table class="data-table">
        <thead>
          <tr>
            <th>Title</th>
            <th>Artist</th>
            <th>Duration</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="song in songs" :key="song.id">
            <td>{{ song.title }}</td>
            <td>{{ song.artist }}</td>
            <td>{{ formatDuration(song.duration) }}</td>
          </tr>
        </tbody>
      </table>

      <Pagination v-model="page" :total="total" :limit="limit" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import Pagination from '../components/Pagination.vue';
import { useApi, type Song } from '../composables/useApi';

const { getMySongs } = useApi();

const songs = ref<Song[]>([]);
const total = ref(0);
const page = ref(1);
const limit = 20;
const loading = ref(false);
const error = ref<string | null>(null);

async function load() {
  loading.value = true;
  error.value = null;
  const { data, error: err } = await getMySongs(page.value, limit);
  if (err || !data) {
    error.value = err ?? 'Failed to load songs';
  } else {
    songs.value = data.data;
    total.value = data.total;
  }
  loading.value = false;
}

function formatDuration(ms: number): string {
  const totalSec = Math.floor(ms / 1000);
  const min = Math.floor(totalSec / 60);
  const sec = totalSec % 60;
  return `${min}:${sec.toString().padStart(2, '0')}`;
}

watch(page, load);
onMounted(load);
</script>
