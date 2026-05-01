<template>
  <div class="page-content">
    <h1 class="page-title">My Songs</h1>

    <SongFilters mode="user" />
    <div class="toolbar">
      <SongSort />
    </div>

    <div v-if="loading" class="state-msg">Loading…</div>
    <div v-else-if="error" class="state-msg error">{{ error }}</div>
    <div v-else-if="songs.length === 0" class="state-msg">No songs match your filters.</div>

    <template v-else>
      <table class="data-table">
        <thead>
          <tr>
            <th>Title</th>
            <th>Artist</th>
            <th>Duration</th>
            <th>Added</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="song in songs" :key="song.id">
            <td>{{ song.title }}</td>
            <td>{{ song.artist }}</td>
            <td>{{ formatDuration(song.duration) }}</td>
            <td>{{ formatDate(song.created_at) }}</td>
          </tr>
        </tbody>
      </table>

      <Pagination :model-value="page" :total="total" :limit="limit" @update:model-value="onPageChange" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import Pagination from '../components/Pagination.vue';
import SongFilters from '../components/SongFilters.vue';
import SongSort from '../components/SongSort.vue';
import { useApi } from '../composables/useApi';
import { useSongFilters } from '../composables/useSongFilters';
import type { SongListItem } from '../types/api';

const { listMySongs } = useApi();
const { filters, update } = useSongFilters();

const songs = ref<SongListItem[]>([]);
const total = ref(0);
const loading = ref(false);
const error = ref<string | null>(null);

const page = computed(() => filters.value.page ?? 1);
const limit = computed(() => filters.value.limit ?? 20);

async function load() {
  loading.value = true;
  error.value = null;
  const { data, error: err } = await listMySongs(filters.value);
  if (err || !data) {
    error.value = err ?? 'Failed to load songs';
  } else {
    songs.value = data.data;
    total.value = data.total;
  }
  loading.value = false;
}

function onPageChange(p: number) {
  update({ page: p });
}

function formatDuration(ms: number): string {
  const totalSec = Math.floor(ms / 1000);
  const min = Math.floor(totalSec / 60);
  const sec = totalSec % 60;
  return `${min}:${sec.toString().padStart(2, '0')}`;
}

function formatDate(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toLocaleString();
}

watch(() => filters.value, load, { deep: true });
onMounted(load);
</script>

<style scoped>
.toolbar { display: flex; justify-content: flex-end; margin: 0.5rem 0 1rem; }
</style>
