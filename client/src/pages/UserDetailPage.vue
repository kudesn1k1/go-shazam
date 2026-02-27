<template>
  <div class="page-content">
    <RouterLink to="/users" class="back-link">← Back to Users</RouterLink>

    <div v-if="loadingUser" class="state-msg">Loading…</div>
    <div v-else-if="userError" class="state-msg error">{{ userError }}</div>

    <template v-else-if="userDetail">
      <h1 class="page-title">{{ userDetail.email }}</h1>
      <p class="meta">Registered: {{ formatDate(userDetail.created_at) }}</p>

      <div class="roles-section">
        <h2>Roles</h2>
        <div class="roles-checkboxes">
          <label v-for="role in availableRoles" :key="role" class="role-check">
            <input
              type="checkbox"
              :value="role"
              v-model="selectedRoles"
            />
            {{ role }}
          </label>
        </div>
        <button class="pill" :disabled="savingRoles" @click="saveRoles">
          {{ savingRoles ? 'Saving…' : 'Save Roles' }}
        </button>
        <span v-if="rolesMsg" class="roles-msg" :class="{ error: rolesError }">{{ rolesMsg }}</span>
      </div>

      <div class="songs-section">
        <h2>Songs</h2>

        <div v-if="loadingSongs" class="state-msg">Loading…</div>
        <div v-else-if="songsError" class="state-msg error">{{ songsError }}</div>
        <div v-else-if="songs.length === 0" class="state-msg">No songs uploaded.</div>

        <template v-else>
          <table class="data-table">
            <thead>
              <tr>
                <th>Title</th>
                <th>Artist</th>
                <th>Duration</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="song in songs" :key="song.id">
                <td>{{ song.title }}</td>
                <td>{{ song.artist }}</td>
                <td>{{ formatDuration(song.duration) }}</td>
              </tr>
            </tbody>
          </table>

          <Pagination v-model="songsPage" :total="songsTotal" :limit="songsLimit" />
        </template>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import { RouterLink, useRoute } from 'vue-router';
import Pagination from '../components/Pagination.vue';
import { useApi, type Song, type UserItem } from '../composables/useApi';

const route = useRoute();
const { getUserByID, getUserSongs, updateUserRoles } = useApi();

const userID = route.params.id as string;
const availableRoles = ['admin', 'user'];

const userDetail = ref<UserItem | null>(null);
const loadingUser = ref(false);
const userError = ref<string | null>(null);

const selectedRoles = ref<string[]>([]);
const savingRoles = ref(false);
const rolesMsg = ref('');
const rolesError = ref(false);

const songs = ref<Song[]>([]);
const songsTotal = ref(0);
const songsPage = ref(1);
const songsLimit = 20;
const loadingSongs = ref(false);
const songsError = ref<string | null>(null);

async function loadUser() {
  loadingUser.value = true;
  userError.value = null;
  const { data, error } = await getUserByID(userID);
  if (error || !data) {
    userError.value = error ?? 'Failed to load user';
  } else {
    userDetail.value = data;
    selectedRoles.value = [...data.roles];
  }
  loadingUser.value = false;
}

async function loadSongs() {
  loadingSongs.value = true;
  songsError.value = null;
  const { data, error } = await getUserSongs(userID, songsPage.value, songsLimit);
  if (error || !data) {
    songsError.value = error ?? 'Failed to load songs';
  } else {
    songs.value = data.data;
    songsTotal.value = data.total;
  }
  loadingSongs.value = false;
}

async function saveRoles() {
  savingRoles.value = true;
  rolesMsg.value = '';
  const { error } = await updateUserRoles(userID, selectedRoles.value);
  if (error) {
    rolesMsg.value = error;
    rolesError.value = true;
  } else {
    rolesMsg.value = 'Roles saved';
    rolesError.value = false;
    if (userDetail.value) {
      userDetail.value.roles = [...selectedRoles.value];
    }
  }
  savingRoles.value = false;
  setTimeout(() => { rolesMsg.value = ''; }, 3000);
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString();
}

function formatDuration(ms: number): string {
  const totalSec = Math.floor(ms / 1000);
  const min = Math.floor(totalSec / 60);
  const sec = totalSec % 60;
  return `${min}:${sec.toString().padStart(2, '0')}`;
}

watch(songsPage, loadSongs);
onMounted(() => {
  loadUser();
  loadSongs();
});
</script>
