<template>
  <div class="avatar-uploader">
    <UserAvatar
      :src="previewUrl ?? currentUrl"
      :email="email"
      size="lg"
    />

    <div v-if="error" class="error">{{ error }}</div>

    <div v-if="state === 'idle' || state === 'error' || state === 'committed'" class="actions">
      <label class="pill">
        Choose photo
        <input type="file" accept="image/jpeg,image/png,image/webp" @change="onPick" />
      </label>
      <button v-if="currentUrl" type="button" class="ghost" @click="onRemove">Remove avatar</button>
    </div>

    <div v-else-if="state === 'uploading'" class="status">Uploading…</div>

    <div v-else-if="state === 'previewing'" class="actions">
      <button type="button" class="pill" @click="onConfirm">Use this photo</button>
      <button type="button" class="ghost" @click="cancel">Cancel</button>
    </div>

    <div v-else-if="state === 'saving'" class="status">Saving…</div>
  </div>
</template>

<script setup lang="ts">
import UserAvatar from './UserAvatar.vue';
import { useAvatarUpload } from '../composables/useAvatarUpload';

const props = defineProps<{
  target: 'self' | { userId: string };
  currentUrl: string | null;
  email?: string;
}>();

const emit = defineEmits<{
  (e: 'changed', url: string | null): void;
}>();

const { state, error, previewUrl, pick, confirm, cancel, remove } = useAvatarUpload({ target: props.target });

function onPick(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  if (file) pick(file);
  input.value = '';
}

async function onConfirm() {
  const ok = await confirm();
  if (ok) emit('changed', previewUrl.value);
}

async function onRemove() {
  const ok = await remove();
  if (ok) emit('changed', null);
}
</script>

<style scoped>
.avatar-uploader { display: flex; flex-direction: column; align-items: center; gap: 0.75rem; }
.actions { display: flex; gap: 0.5rem; flex-wrap: wrap; justify-content: center; }
.actions input[type=file] { display: none; }
.status { opacity: 0.7; font-size: 0.9rem; }
.error { color: #ff9f9f; font-size: 0.85rem; text-align: center; max-width: 280px; }
</style>
