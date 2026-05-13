<template>
  <div class="page-content">
    <h1 class="page-title">All Songs</h1>

    <SongFilters mode="admin" />
    <div class="toolbar">
      <SongSort />
    </div>

    <div v-if="loading" class="state-msg">Loading…</div>
    <div v-else-if="error" class="state-msg error">{{ error }}</div>
    <div v-else-if="songs.length === 0" class="state-msg">No songs found.</div>

    <template v-else>
      <table class="data-table">
        <thead>
          <tr>
            <th>Title</th>
            <th>Artist</th>
            <th>Duration</th>
            <th>Added</th>
            <th>Source</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="s in songs" :key="s.id">
            <td>{{ s.title }}</td>
            <td>{{ s.artist }}</td>
            <td>{{ formatDuration(s.duration) }}</td>
            <td>{{ formatDate(s.created_at) }}</td>
            <td>
              <a :href="`https://www.youtube.com/watch?v=${s.source_id}`" target="_blank" class="source-link">
                YouTube
              </a>
            </td>
            <td>
              <button class="ghost danger-btn" @click="handleDelete(s.id)" :disabled="deleting === s.id">
                {{ deleting === s.id ? 'Deleting...' : 'Delete' }}
              </button>
            </td>
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
import { useToast } from '../composables/useToast';
import { useNoindex } from '../composables/useSeo';
import type { SongListItem } from '../types/api';

useNoindex('All Songs');

const { listAllSongs, deleteSong } = useApi();
const { filters, update } = useSongFilters();
const toast = useToast();

const songs = ref<SongListItem[]>([]);
const total = ref(0);
const loading = ref(false);
const error = ref<string | null>(null);
const deleting = ref<string | null>(null);

const page = computed(() => filters.value.page ?? 1);
const limit = computed(() => filters.value.limit ?? 20);

async function load() {
  loading.value = true;
  error.value = null;
  const { data, error: err } = await listAllSongs(filters.value);
  if (err || !data) {
    error.value = err ?? 'Failed to load songs';
  } else {
    songs.value = data.data;
    total.value = data.total;
  }
  loading.value = false;
}

async function handleDelete(id: string) {
  if (!confirm('Are you sure you want to delete this song? This will also remove all associated fingerprints.')) {
    return;
  }

  deleting.value = id;
  const { error: err } = await deleteSong(id);
  if (err) {
    toast.show(err, 'error');
  } else {
    toast.show('Song deleted successfully', 'success');
    await load();
  }
  deleting.value = null;
}

function onPageChange(p: number) {
  update({ page: p });
}

function formatDuration(ms: number): string {
  const seconds = Math.floor(ms / 1000);
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  return `${m}:${s.toString().padStart(2, '0')}`;
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
.source-link {
  color: #3ea6ff;
  text-decoration: none;
  font-size: 0.9rem;
}

.source-link:hover {
  text-decoration: underline;
}

.danger-btn {
  color: #ff4e4e !important;
}

.danger-btn:hover:not(:disabled) {
  background: rgba(255, 78, 78, 0.1) !important;
}

.danger-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
