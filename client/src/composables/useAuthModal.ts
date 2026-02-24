import { ref } from 'vue';

const isOpen = ref(false);
const mode = ref<'login' | 'register'>('login');

export function useAuthModal() {
  function open(m: 'login' | 'register' = 'login') {
    mode.value = m;
    isOpen.value = true;
  }

  function close() {
    isOpen.value = false;
  }

  return { isOpen, mode, open, close };
}
