import { computed, ref } from 'vue';
import { useApi } from './useApi';
import { MAX_AVATAR_BYTES } from '../types/api';

type State = 'idle' | 'uploading' | 'previewing' | 'saving' | 'committed' | 'error';

const ALLOWED_EXTENSIONS = ['.jpg', '.jpeg', '.png', '.webp'];

function extensionOf(name: string): string {
  const i = name.lastIndexOf('.');
  return i >= 0 ? name.slice(i).toLowerCase() : '';
}

export function useAvatarUpload(opts: { target: 'self' | { userId: string } }) {
  const api = useApi();

  const state = ref<State>('idle');
  const error = ref<string | null>(null);
  const previewFile = ref<File | null>(null);
  const previewUrl = ref<string | null>(null);
  const uploadedHash = ref<string | null>(null);
  const committedUrl = ref<string | null>(null);
  const progress = ref(0);

  const canPick = computed(() => state.value === 'idle' || state.value === 'error' || state.value === 'committed');

  async function pick(file: File) {
    error.value = null;

    if (!ALLOWED_EXTENSIONS.includes(extensionOf(file.name))) {
      state.value = 'error';
      error.value = 'Unsupported format (use JPG, PNG, or WebP)';
      return;
    }
    if (file.size > MAX_AVATAR_BYTES) {
      state.value = 'error';
      error.value = 'File too large (max 2 MB)';
      return;
    }

    state.value = 'uploading';
    progress.value = 0;

    const { data, error: err, status } = await api.uploadFile(file);
    if (err || !data) {
      state.value = 'error';
      error.value = err ?? `Upload failed (status ${status})`;
      return;
    }

    uploadedHash.value = data.hash;
    previewFile.value = file;
    if (previewUrl.value) URL.revokeObjectURL(previewUrl.value);
    previewUrl.value = URL.createObjectURL(file);
    state.value = 'previewing';
  }

  async function confirm(): Promise<boolean> {
    if (!uploadedHash.value) return false;
    state.value = 'saving';
    error.value = null;

    const req = opts.target === 'self'
      ? api.setOwnAvatar(uploadedHash.value)
      : api.setUserAvatar(opts.target.userId, uploadedHash.value);
    const { data, error: err } = await req;

    if (err || !data) {
      state.value = 'error';
      error.value = err ?? 'Save failed';
      return false;
    }

    committedUrl.value = data.avatar_url;
    state.value = 'committed';
    return true;
  }

  function cancel() {
    if (previewUrl.value) URL.revokeObjectURL(previewUrl.value);
    previewUrl.value = null;
    previewFile.value = null;
    uploadedHash.value = null;
    error.value = null;
    state.value = 'idle';
  }

  async function remove(): Promise<boolean> {
    error.value = null;
    const { error: err } = opts.target === 'self'
      ? await api.clearOwnAvatar()
      : await api.clearUserAvatar(opts.target.userId);
    if (err) {
      error.value = err;
      return false;
    }
    committedUrl.value = null;
    return true;
  }

  return {
    state, error, previewUrl, progress, committedUrl, canPick,
    pick, confirm, cancel, remove,
  };
}
