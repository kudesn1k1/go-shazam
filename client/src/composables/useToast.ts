import { ref } from 'vue';

const message = ref('');
const type = ref<'success' | 'error'>('success');
let timer: ReturnType<typeof setTimeout> | null = null;

export function useToast() {
  function show(msg: string, t: 'success' | 'error' = 'success') {
    if (timer) clearTimeout(timer);
    message.value = msg;
    type.value = t;
    timer = setTimeout(() => { message.value = ''; }, 4000);
  }

  return { message, type, show };
}
