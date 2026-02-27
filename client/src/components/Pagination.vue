<template>
  <div v-if="totalPages > 1" class="pagination">
    <button
      class="page-btn"
      :disabled="modelValue <= 1"
      @click="emit('update:modelValue', modelValue - 1)"
    >
      ‹
    </button>

    <button
      v-for="page in visiblePages"
      :key="page"
      class="page-btn"
      :class="{ active: page === modelValue }"
      @click="emit('update:modelValue', page)"
    >
      {{ page }}
    </button>

    <button
      class="page-btn"
      :disabled="modelValue >= totalPages"
      @click="emit('update:modelValue', modelValue + 1)"
    >
      ›
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{
  modelValue: number;
  total: number;
  limit: number;
}>();

const emit = defineEmits<{
  'update:modelValue': [page: number];
}>();

const totalPages = computed(() => Math.ceil(props.total / props.limit));

const visiblePages = computed(() => {
  const pages: number[] = [];
  const current = props.modelValue;
  const total = totalPages.value;
  const delta = 2;

  for (let i = Math.max(1, current - delta); i <= Math.min(total, current + delta); i++) {
    pages.push(i);
  }

  return pages;
});
</script>

<style scoped>
.pagination {
  display: flex;
  gap: 0.25rem;
  justify-content: center;
  margin-top: 1.5rem;
}

.page-btn {
  min-width: 2rem;
  height: 2rem;
  padding: 0 0.5rem;
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.05);
  color: rgba(255, 255, 255, 0.7);
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.15s;
}

.page-btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.page-btn.active {
  background: rgba(255, 78, 126, 0.3);
  border-color: rgba(255, 78, 126, 0.5);
  color: #fff;
}

.page-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}
</style>
