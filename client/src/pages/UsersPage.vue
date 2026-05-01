<template>
  <div class="page-content">
    <h1 class="page-title">Users</h1>

    <div v-if="loading" class="state-msg">Loading…</div>
    <div v-else-if="error" class="state-msg error">{{ error }}</div>
    <div v-else-if="users.length === 0" class="state-msg">No users found.</div>

    <template v-else>
      <table class="data-table">
        <thead>
          <tr>
            <th></th>
            <th>Email</th>
            <th>Roles</th>
            <th>Registered</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id">
            <td><UserAvatar :src="u.avatar_url" :email="u.email" size="sm" /></td>
            <td>{{ u.email }}</td>
            <td>
              <span v-for="role in u.roles" :key="role" class="role-badge" :class="role">
                {{ role }}
              </span>
            </td>
            <td>{{ formatDate(u.created_at) }}</td>
            <td>
              <RouterLink :to="`/users/${u.id}`" class="link-btn">View</RouterLink>
            </td>
          </tr>
        </tbody>
      </table>

      <Pagination v-model="page" :total="total" :limit="limit" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import { RouterLink } from 'vue-router';
import Pagination from '../components/Pagination.vue';
import UserAvatar from '../components/UserAvatar.vue';
import { useApi } from '../composables/useApi';
import type { UserWithAvatar } from '../types/api';

const { listUsers } = useApi();

const users = ref<UserWithAvatar[]>([]);
const total = ref(0);
const page = ref(1);
const limit = 20;
const loading = ref(false);
const error = ref<string | null>(null);

async function load() {
  loading.value = true;
  error.value = null;
  const { data, error: err } = await listUsers(page.value, limit);
  if (err || !data) {
    error.value = err ?? 'Failed to load users';
  } else {
    users.value = data.data;
    total.value = data.total;
  }
  loading.value = false;
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString();
}

watch(page, load);
onMounted(load);
</script>
