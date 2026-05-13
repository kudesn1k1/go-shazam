<template>
  <div :class="['avatar', sizeClass]" :style="{ backgroundColor: color }">
    <img
      v-if="displaySrc"
      :src="displaySrc"
      :alt="alt"
      :width="pixelSize"
      :height="pixelSize"
      loading="lazy"
      decoding="async"
      @error="onImgError"
    />
    <span v-else class="initials">{{ initials }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { API_BASE_URL } from '../composables/useApi';

const props = withDefaults(
  defineProps<{
    src: string | null | undefined;
    email?: string;
    size?: 'sm' | 'md' | 'lg';
    alt?: string;
  }>(),
  { size: 'md', alt: 'avatar' }
);

const sizeClass = computed(() => `size-${props.size}`);
const pixelSize = computed(() => ({ sm: 32, md: 64, lg: 128 } as const)[props.size]);

const imgErrored = ref(false);
watch(() => props.src, () => { imgErrored.value = false; });
function onImgError() { imgErrored.value = true; }

const displaySrc = computed(() => {
  if (imgErrored.value || !props.src) return null;
  // Relative /api/* paths need the base URL prefix during dev so Vite doesn't intercept them.
  if (props.src.startsWith('/')) return `${API_BASE_URL}${props.src}`;
  return props.src;
});

const initials = computed(() => {
  const email = props.email ?? '';
  if (!email) return '?';
  const local = email.split('@')[0];
  if (!local || local.length === 0) return '?';
  return local.slice(0, 2).toUpperCase();
});

const color = computed(() => {
  const email = props.email ?? '';
  let hash = 0;
  for (let i = 0; i < email.length; i++) hash = (hash * 31 + email.charCodeAt(i)) >>> 0;
  const hue = hash % 360;
  return `hsl(${hue}, 45%, 32%)`;
});
</script>

<style scoped>
.avatar { display: inline-flex; align-items: center; justify-content: center; border-radius: 50%; overflow: hidden; color: #fff; font-weight: 600; }
.avatar img { width: 100%; height: 100%; object-fit: cover; display: block; }
.size-sm { width: 32px; height: 32px; font-size: 0.75rem; }
.size-md { width: 64px; height: 64px; font-size: 1rem; }
.size-lg { width: 128px; height: 128px; font-size: 1.75rem; }
</style>
