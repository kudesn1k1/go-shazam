<template>
  <section class="profile">
    <h1>Profile</h1>
    <p v-if="user" class="email">{{ user.email }}</p>
    <AvatarUploader
      v-if="user"
      target="self"
      :current-url="user.avatar_url"
      :email="user.email"
      @changed="onAvatarChanged"
    />
  </section>
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import { useAuth } from '../composables/useAuth';
import { useNoindex } from '../composables/useSeo';
import AvatarUploader from '../components/AvatarUploader.vue';

useNoindex('Profile');

const { user, fetchUser } = useAuth();

onMounted(() => {
  if (!user.value) fetchUser();
});

async function onAvatarChanged() {
  await fetchUser();
}
</script>

<style scoped>
.profile {
  max-width: 480px;
  margin: 2rem auto;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  align-items: center;
}
.email { opacity: 0.75; }
</style>
