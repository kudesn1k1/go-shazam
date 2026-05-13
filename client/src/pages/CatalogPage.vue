<template>
  <main class="page-content">
    <header class="catalog-header">
      <h1 class="page-title">Music catalog</h1>
      <p class="lead">
        Browse the songs we already recognize. Tap any track to see details, or
        head back to the home page to identify whatever's playing around you.
      </p>
    </header>

    <SongFilters mode="user" />
    <div class="toolbar">
      <SongSort />
    </div>

    <div v-if="loading" class="state-msg" role="status">Loading catalog…</div>
    <div v-else-if="error" class="state-msg error" role="alert">{{ error }}</div>
    <div v-else-if="songs.length === 0" class="state-msg">No songs match your filters.</div>

    <template v-else>
      <table class="data-table">
        <thead>
          <tr>
            <th scope="col">Title</th>
            <th scope="col">Artist</th>
            <th scope="col">Duration</th>
            <th scope="col">Added</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="song in songs" :key="song.id">
            <td>
              <RouterLink :to="`/catalog/${song.id}`" class="song-link">
                {{ song.title }}
              </RouterLink>
            </td>
            <td>{{ song.artist }}</td>
            <td>{{ formatDuration(song.duration) }}</td>
            <td>{{ formatDate(song.created_at) }}</td>
          </tr>
        </tbody>
      </table>

      <Pagination :model-value="page" :total="total" :limit="limit" @update:model-value="onPageChange" />
    </template>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { RouterLink } from 'vue-router';
import { useHead } from '@unhead/vue';
import Pagination from '../components/Pagination.vue';
import SongFilters from '../components/SongFilters.vue';
import SongSort from '../components/SongSort.vue';
import { useApi } from '../composables/useApi';
import { useSongFilters } from '../composables/useSongFilters';
import type { PublicSong } from '../types/api';

useHead({
  title: 'Music catalog — Go Shazam',
  meta: [
    { name: 'description', content: 'Browse the songs Go Shazam already recognizes. Search by title or artist, sort by recency, duration, or alphabetical order.' },
    { property: 'og:title', content: 'Music catalog — Go Shazam' },
    { property: 'og:description', content: 'Browse the songs Go Shazam already recognizes.' },
    { property: 'og:type', content: 'website' },
  ],
  link: [
    { rel: 'canonical', href: () => (typeof window !== 'undefined' ? `${window.location.origin}/catalog` : '/catalog') },
  ],
});

const { listPublicSongs } = useApi();
const { filters, update } = useSongFilters();

const songs = ref<PublicSong[]>([]);
const total = ref(0);
const loading = ref(false);
const error = ref<string | null>(null);

const page = computed(() => filters.value.page ?? 1);
const limit = computed(() => filters.value.limit ?? 20);

async function load() {
  loading.value = true;
  error.value = null;
  const { data, error: err } = await listPublicSongs(filters.value);
  if (err || !data) {
    error.value = err ?? 'Failed to load catalog';
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
.catalog-header { margin-bottom: 1.5rem; }
.lead { color: rgba(255,255,255,0.7); max-width: 60ch; line-height: 1.5; }
.toolbar { display: flex; justify-content: flex-end; margin: 0.5rem 0 1rem; }
.song-link { color: inherit; text-decoration: none; font-weight: 500; }
.song-link:hover { text-decoration: underline; }
</style>
