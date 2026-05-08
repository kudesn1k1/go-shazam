<template>
  <main class="page-content song-detail">
    <RouterLink to="/catalog" class="back-link">← Back to catalog</RouterLink>

    <div v-if="loading" class="state-msg" role="status">Loading song…</div>

    <article v-else-if="song" class="song-card">
      <header>
        <h1 class="song-title">{{ song.title }}</h1>
        <h2 class="song-artist">{{ song.artist }}</h2>
      </header>

      <dl class="meta">
        <div class="meta-row">
          <dt>Duration</dt>
          <dd>{{ formatDuration(song.duration) }}</dd>
        </div>
        <div class="meta-row">
          <dt>Added</dt>
          <dd>
            <time :datetime="song.created_at">{{ formatDate(song.created_at) }}</time>
          </dd>
        </div>
      </dl>

      <div class="actions">
        <a
          v-if="spotifyUrl"
          :href="spotifyUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="pill"
        >
          Open on Spotify
        </a>
        <RouterLink to="/" class="ghost">Try recognizing music</RouterLink>
      </div>
    </article>

    <div v-else-if="notFound" class="state-msg" role="alert">
      <h1>Song not found</h1>
      <p>This song isn't in our catalog. It may have been removed.</p>
      <RouterLink to="/catalog" class="pill">Back to catalog</RouterLink>
    </div>

    <div v-else-if="error" class="state-msg error" role="alert">
      <p>{{ error }}</p>
      <button class="ghost" @click="load">Try again</button>
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { RouterLink, useRoute } from 'vue-router';
import { useHead } from '@unhead/vue';
import { useApi } from '../composables/useApi';
import type { PublicSong } from '../types/api';

const route = useRoute();
const { getPublicSong } = useApi();

const song = ref<PublicSong | null>(null);
const loading = ref(false);
const error = ref<string | null>(null);
const notFound = ref(false);

const pageTitle = computed(() => {
  if (notFound.value) return 'Song not found — Go Shazam';
  if (!song.value) return 'Loading… — Go Shazam';
  return `${song.value.title} by ${song.value.artist} — Go Shazam`;
});

const pageDescription = computed(() => {
  if (!song.value) return 'Song detail page on Go Shazam.';
  return `Listen to "${song.value.title}" by ${song.value.artist}. Part of the Go Shazam catalog.`;
});

const canonicalHref = computed(() => {
  if (typeof window === 'undefined') return `/catalog/${route.params.id}`;
  return `${window.location.origin}/catalog/${route.params.id}`;
});

useHead({
  title: pageTitle,
  meta: [
    { name: 'description', content: pageDescription },
    { name: 'robots', content: () => (notFound.value ? 'noindex' : 'index, follow') },
    { property: 'og:title', content: pageTitle },
    { property: 'og:description', content: pageDescription },
    { property: 'og:type', content: 'music.song' },
    { property: 'og:url', content: canonicalHref },
  ],
  link: [
    { rel: 'canonical', href: canonicalHref },
  ],
});

const spotifyUrl = computed(() => {
  if (!song.value?.source_id) return null;
  // SourceID stored in DB is the platform-native ID — for Spotify-ingested
  // songs it's the Spotify track ID. yt-dlp-ingested songs may use other IDs;
  // detect a Spotify-shaped 22-char base62 ID before linking.
  if (/^[A-Za-z0-9]{22}$/.test(song.value.source_id)) {
    return `https://open.spotify.com/track/${song.value.source_id}`;
  }
  return null;
});

async function load() {
  const id = route.params.id as string;
  if (!id) return;

  loading.value = true;
  error.value = null;
  notFound.value = false;
  song.value = null;

  const { data, error: err, status } = await getPublicSong(id);
  if (status === 404) {
    notFound.value = true;
  } else if (err || !data) {
    error.value = err ?? 'Failed to load song';
  } else {
    song.value = data;
  }
  loading.value = false;
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

watch(() => route.params.id, load);
onMounted(load);
</script>

<style scoped>
.song-detail { max-width: 720px; }
.back-link { display: inline-block; margin-bottom: 1.5rem; color: rgba(255,255,255,0.7); text-decoration: none; font-size: 0.9rem; }
.back-link:hover { color: #fff; }
.song-card { background: rgba(255,255,255,0.04); border: 1px solid rgba(255,255,255,0.08); border-radius: 12px; padding: 2rem; }
.song-title { font-size: 2.5rem; margin: 0 0 0.5rem; }
.song-artist { font-size: 1.25rem; font-weight: 500; color: rgba(255,255,255,0.75); margin: 0 0 1.5rem; }
.meta { display: grid; gap: 0.5rem; margin: 1.5rem 0; }
.meta-row { display: flex; gap: 1rem; }
.meta-row dt { color: rgba(255,255,255,0.55); min-width: 80px; font-size: 0.9rem; }
.meta-row dd { margin: 0; font-size: 0.95rem; }
.actions { display: flex; gap: 0.75rem; flex-wrap: wrap; margin-top: 1.5rem; }
</style>
