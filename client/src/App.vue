<template>
  <div class="page">
    <header class="top-bar">
      <div class="branding">
        <span class="pulse-dot" aria-hidden="true" />
        <div>
          <RouterLink to="/" class="brand-link">
            <strong>Go Shazam</strong>
          </RouterLink>
          <p>Find the music around you</p>
        </div>
      </div>

      <nav v-if="isAuthenticated" class="nav-links">
        <RouterLink to="/my-songs" class="nav-link">My Songs</RouterLink>
        <RouterLink v-if="isAdmin" to="/songs" class="nav-link">All Songs</RouterLink>
        <RouterLink v-if="isAdmin" to="/users" class="nav-link">Users</RouterLink>
      </nav>

      <div class="actions">
        <template v-if="isAuthenticated">
          <span class="user-email">{{ user?.email }}</span>
          <button class="ghost" @click="handleLogout">Log out</button>
        </template>
        <template v-else>
          <button class="ghost" @click="authModal.open('login')">Log in</button>
          <button class="pill" @click="authModal.open('register')">Create account</button>
        </template>
      </div>
    </header>

    <RouterView />

    <AuthModal
      v-model="authModal.isOpen.value"
      :mode="authModal.mode.value"
      :pending="authLoading"
      @submit="handleAuthSubmit"
    />

    <transition name="toast">
      <div v-if="toast.message.value" class="toast" :class="{ 'toast-error': toast.type.value === 'error' }">
        {{ toast.message.value }}
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import { RouterLink, RouterView, useRouter } from 'vue-router';

import AuthModal from './components/AuthModal.vue';
import { useAuth } from './composables/useAuth';
import { useAuthModal } from './composables/useAuthModal';
import { useToast } from './composables/useToast';

const router = useRouter();
const authModal = useAuthModal();
const toast = useToast();

const {
  user,
  isAuthenticated,
  isAdmin,
  isLoading: authLoading,
  error: authError,
  login,
  register,
  logout,
  initialize: initAuth,
} = useAuth();

onMounted(async () => {
  await initAuth();
});

const handleAuthSubmit = async (payload: { mode: 'login' | 'register'; email: string; password: string }) => {
  const success = payload.mode === 'login'
    ? await login(payload.email, payload.password)
    : await register(payload.email, payload.password);

  if (success) {
    authModal.close();
    toast.show(
      payload.mode === 'login' ? 'Welcome back!' : 'Account created successfully!',
      'success'
    );
  } else {
    toast.show(authError.value || 'An error occurred', 'error');
  }
};

const handleLogout = async () => {
  await logout();
  router.push('/');
  toast.show('You have been logged out', 'success');
};
</script>

<style scoped>
.brand-link {
  text-decoration: none;
  color: inherit;
}

.nav-links {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.nav-link {
  color: rgba(255, 255, 255, 0.7);
  text-decoration: none;
  font-size: 0.95rem;
  font-weight: 500;
  transition: color 0.15s;
}

.nav-link:hover,
.nav-link.router-link-active {
  color: #fff;
}

.user-email {
  font-size: 0.9rem;
  opacity: 0.8;
  margin-right: 0.5rem;
  display: flex;
  align-items: center;
}

.toast-error {
  background: rgba(255, 78, 78, 0.15) !important;
  border-color: rgba(255, 78, 78, 0.3) !important;
  color: #ff9f9f !important;
}
</style>
