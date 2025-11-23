<template>
  <Teleport to="body">
    <div v-if="modelValue" class="modal-overlay" @click.self="close">
      <section class="modal">
        <header>
          <h2>{{ title }}</h2>
          <button class="close-btn" @click="close" aria-label="Close dialog">
            ×
          </button>
        </header>

        <form @submit.prevent="handleSubmit">
          <div v-if="isRegister" class="form-group">
            <label for="name">Name</label>
            <input id="name" v-model="form.name" type="text" placeholder="Jane Doe" autocomplete="name" />
          </div>
          <div class="form-group">
            <label for="email">Email</label>
            <input
              id="email"
              v-model="form.email"
              type="email"
              placeholder="you@example.com"
              autocomplete="email"
              required
            />
          </div>
          <div class="form-group">
            <label for="password">Password</label>
            <input
              id="password"
              v-model="form.password"
              type="password"
              placeholder="********"
              autocomplete="current-password"
              required
            />
          </div>
          <div v-if="isRegister" class="form-group">
            <label for="confirm">Confirm password</label>
            <input
              id="confirm"
              v-model="form.confirm"
              type="password"
              placeholder="********"
              autocomplete="new-password"
              required
            />
          </div>

          <button class="primary" type="submit" :disabled="pending">
            {{ pending ? 'Please wait…' : submitLabel }}
          </button>
        </form>
      </section>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, reactive, watch } from 'vue';

type AuthMode = 'login' | 'register';

const props = defineProps<{
  modelValue: boolean;
  mode: AuthMode;
  pending?: boolean;
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void;
  (e: 'submit', payload: { mode: AuthMode; email: string; password: string; name?: string }): void;
}>();

const form = reactive({
  name: '',
  email: '',
  password: '',
  confirm: '',
});

watch(
  () => props.mode,
  () => {
    form.name = '';
    form.email = '';
    form.password = '';
    form.confirm = '';
  }
);

const title = computed(() => (props.mode === 'login' ? 'Sign in to continue' : 'Create your account'));

const submitLabel = computed(() => (props.mode === 'login' ? 'Sign in' : 'Create account'));

const isRegister = computed(() => props.mode === 'register');

const close = () => emit('update:modelValue', false);

const handleSubmit = () => {
  if (!form.email || !form.password) {
    return;
  }

  if (isRegister.value && form.password !== form.confirm) {
    return;
  }

  emit('submit', {
    mode: props.mode,
    email: form.email,
    password: form.password,
    name: form.name || undefined,
  });
};
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(4, 8, 25, 0.75);
  backdrop-filter: blur(4px);
  display: grid;
  place-items: center;
  z-index: 50;
}

.modal {
  width: min(420px, 92%);
  background: #0f1220;
  border-radius: 18px;
  padding: 24px;
  color: #fff;
  box-shadow: 0 24px 60px rgba(4, 8, 25, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.06);
}

header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
}

h2 {
  font-size: 1.35rem;
  margin: 0;
}

.close-btn {
  background: transparent;
  border: none;
  color: #fff;
  font-size: 1.5rem;
  cursor: pointer;
  line-height: 1;
}

form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
}

label {
  font-size: 0.85rem;
  margin-bottom: 0.35rem;
  opacity: 0.8;
}

input {
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.15);
  background: rgba(255, 255, 255, 0.08);
  padding: 0.75rem;
  color: #fff;
  font-size: 0.95rem;
}

input:focus {
  outline: 2px solid rgba(0, 173, 255, 0.7);
}

button.primary {
  margin-top: 0.5rem;
  border: none;
  border-radius: 999px;
  padding: 0.85rem;
  font-size: 1rem;
  font-weight: 600;
  background: linear-gradient(135deg, #009dff, #6a5af9);
  color: #fff;
  cursor: pointer;
  transition: opacity 0.2s ease;
}

button.primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>

