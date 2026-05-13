<template>
  <div class="sort">
    <label>
      <span>Sort</span>
      <select :value="filters.sort" @change="onSort(($event.target as HTMLSelectElement).value)">
        <option value="created_at">Date added</option>
        <option value="title">Title</option>
        <option value="artist">Artist</option>
        <option value="duration">Duration</option>
      </select>
    </label>
    <button type="button" class="ghost" @click="toggleOrder">
      {{ filters.order === 'asc' ? '↑ Asc' : '↓ Desc' }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { useSongFilters } from '../composables/useSongFilters';
import type { SortField } from '../types/api';

const { filters, update } = useSongFilters();

function onSort(value: string) {
  update({ sort: value as SortField });
}

function toggleOrder() {
  update({ order: filters.value.order === 'asc' ? 'desc' : 'asc' });
}
</script>

<style scoped>
.sort { display: inline-flex; gap: 0.5rem; align-items: end; }
.sort label { display: flex; flex-direction: column; font-size: 0.85rem; gap: 0.25rem; }
.sort select { padding: 0.4rem 0.6rem; border-radius: 6px; border: 1px solid rgba(255,255,255,0.15); background: #1a1f2e; color: #fff; }
.sort select option { background: #1a1f2e; color: #fff; }
</style>
