import { computed, type ComputedRef } from 'vue';
import { useRoute, useRouter, type LocationQueryRaw } from 'vue-router';
import type { SongFilterQuery, SortField, SortOrder } from '../types/api';

const VALID_SORT: SortField[] = ['title', 'artist', 'duration', 'created_at'];
const VALID_ORDER: SortOrder[] = ['asc', 'desc'];

function pickSort(raw: unknown): SortField {
  return VALID_SORT.includes(raw as SortField) ? (raw as SortField) : 'created_at';
}
function pickOrder(raw: unknown): SortOrder {
  return VALID_ORDER.includes(raw as SortOrder) ? (raw as SortOrder) : 'desc';
}
function pickInt(raw: unknown, fallback: number): number {
  const n = Number(raw);
  return Number.isFinite(n) && n > 0 ? Math.floor(n) : fallback;
}
function pickString(raw: unknown): string | undefined {
  if (typeof raw !== 'string') return undefined;
  const trimmed = raw.trim();
  return trimmed === '' ? undefined : trimmed;
}

export function useSongFilters(): {
  filters: ComputedRef<SongFilterQuery>;
  update: (patch: Partial<SongFilterQuery>) => void;
  clear: () => void;
} {
  const route = useRoute();
  const router = useRouter();

  const filters = computed<SongFilterQuery>(() => ({
    q:              pickString(route.query.q),
    artist:         pickString(route.query.artist),
    uploaded_by:    pickString(route.query.uploaded_by),
    created_after:  pickString(route.query.created_after),
    created_before: pickString(route.query.created_before),
    sort:           pickSort(route.query.sort),
    order:          pickOrder(route.query.order),
    page:           pickInt(route.query.page, 1),
    limit:          pickInt(route.query.limit, 20),
  }));

  function update(patch: Partial<SongFilterQuery>) {
    const resetPage = !('page' in patch);
    const next: LocationQueryRaw = { ...route.query };
    for (const [k, v] of Object.entries(patch)) {
      if (v === undefined || v === null || v === '') {
        delete (next as Record<string, unknown>)[k];
      } else {
        (next as Record<string, unknown>)[k] = String(v);
      }
    }
    if (resetPage) {
      (next as Record<string, unknown>).page = '1';
    }
    router.push({ path: route.path, query: next });
  }

  function clear() {
    router.push({ path: route.path, query: {} });
  }

  return { filters, update, clear };
}
