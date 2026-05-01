<template>
  <form class="filters" @submit.prevent>
    <div class="row">
      <label class="field">
        <span>Search title or artist</span>
        <input v-model="qLocal" type="text" maxlength="200" placeholder="nirvana, muse…" />
      </label>
      <label class="field">
        <span>Artist</span>
        <input v-model="artistLocal" type="text" maxlength="200" placeholder="exact artist" />
      </label>
    </div>

    <div class="row">
      <label class="field">
        <span>Added after</span>
        <input v-model="afterLocal" type="datetime-local" />
      </label>
      <label class="field">
        <span>Added before</span>
        <input v-model="beforeLocal" type="datetime-local" />
      </label>
    </div>

    <div v-if="mode === 'admin'" class="row">
      <label class="field">
        <span>Uploader user ID</span>
        <input v-model="uploaderLocal" type="text" placeholder="UUID" />
      </label>
    </div>

    <div class="row actions">
      <button type="button" class="ghost" @click="clear">Clear filters</button>
      <div v-if="activeChips.length" class="chips">
        <span v-for="c in activeChips" :key="c.key" class="chip">
          {{ c.label }} <button type="button" @click="removeFilter(c.key)">✕</button>
        </span>
      </div>
    </div>
  </form>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useSongFilters } from '../composables/useSongFilters';
import type { SongFilterQuery } from '../types/api';

defineProps<{ mode: 'user' | 'admin' | 'by-id' }>();
const { filters, update, clear } = useSongFilters();

const qLocal        = ref(filters.value.q ?? '');
const artistLocal   = ref(filters.value.artist ?? '');
const uploaderLocal = ref(filters.value.uploaded_by ?? '');
const afterLocal    = ref(toLocalInput(filters.value.created_after));
const beforeLocal   = ref(toLocalInput(filters.value.created_before));

watch(() => filters.value, (f) => {
  qLocal.value        = f.q ?? '';
  artistLocal.value   = f.artist ?? '';
  uploaderLocal.value = f.uploaded_by ?? '';
  afterLocal.value    = toLocalInput(f.created_after);
  beforeLocal.value   = toLocalInput(f.created_before);
});

const DEBOUNCE_MS = 300;
let qTimer: number | undefined, artistTimer: number | undefined, uploaderTimer: number | undefined;

watch(qLocal, (v) => {
  window.clearTimeout(qTimer);
  qTimer = window.setTimeout(() => update({ q: v || undefined }), DEBOUNCE_MS);
});

watch(artistLocal, (v) => {
  window.clearTimeout(artistTimer);
  artistTimer = window.setTimeout(() => update({ artist: v || undefined }), DEBOUNCE_MS);
});

watch(uploaderLocal, (v) => {
  window.clearTimeout(uploaderTimer);
  uploaderTimer = window.setTimeout(() => update({ uploaded_by: v || undefined }), DEBOUNCE_MS);
});

watch(afterLocal,  (v) => update({ created_after:  v ? localInputToISO(v) : undefined }));
watch(beforeLocal, (v) => update({ created_before: v ? localInputToISO(v) : undefined }));

function toLocalInput(iso: string | undefined): string {
  if (!iso) return '';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '';
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

function localInputToISO(local: string): string | undefined {
  const d = new Date(local);
  if (Number.isNaN(d.getTime())) return undefined;
  return d.toISOString();
}

const activeChips = computed(() => {
  const f = filters.value;
  const chips: Array<{ key: keyof SongFilterQuery; label: string }> = [];
  if (f.q)              chips.push({ key: 'q',              label: `Search: ${f.q}` });
  if (f.artist)         chips.push({ key: 'artist',         label: `Artist: ${f.artist}` });
  if (f.uploaded_by)    chips.push({ key: 'uploaded_by',    label: `Uploader: ${f.uploaded_by.slice(0, 8)}…` });
  if (f.created_after)  chips.push({ key: 'created_after',  label: `After: ${f.created_after.slice(0, 16)}` });
  if (f.created_before) chips.push({ key: 'created_before', label: `Before: ${f.created_before.slice(0, 16)}` });
  return chips;
});

function removeFilter(key: keyof SongFilterQuery) {
  update({ [key]: undefined });
}
</script>

<style scoped>
.filters { display: flex; flex-direction: column; gap: 0.75rem; }
.row { display: flex; gap: 1rem; flex-wrap: wrap; }
.field { display: flex; flex-direction: column; font-size: 0.85rem; gap: 0.25rem; min-width: 180px; }
.field input { padding: 0.4rem 0.6rem; border-radius: 6px; border: 1px solid rgba(255,255,255,0.15); background: rgba(255,255,255,0.04); color: inherit; }
.actions { align-items: center; }
.chips { display: flex; gap: 0.4rem; flex-wrap: wrap; }
.chip { background: rgba(255,255,255,0.08); padding: 0.2rem 0.5rem; border-radius: 999px; font-size: 0.8rem; display: inline-flex; align-items: center; gap: 0.3rem; }
.chip button { background: none; border: none; color: inherit; cursor: pointer; padding: 0; font-size: 0.85rem; }
</style>
