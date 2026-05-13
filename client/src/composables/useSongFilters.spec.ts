import { describe, it, expect, beforeEach, vi } from 'vitest';
import { ref } from 'vue';

// useSongFilters reads from `useRoute()` and writes via `useRouter().push()`.
// We mock vue-router to control route.query and capture push calls.

const currentQuery = ref<Record<string, string>>({});
const pushSpy = vi.fn((to: { path: string; query: Record<string, string> }) => {
  currentQuery.value = to.query;
  return Promise.resolve();
});

vi.mock('vue-router', () => ({
  useRoute: () => ({
    path: '/catalog',
    query: currentQuery.value,
  }),
  useRouter: () => ({
    push: pushSpy,
  }),
}));

describe('useSongFilters', () => {
  beforeEach(() => {
    currentQuery.value = {};
    pushSpy.mockClear();
  });

  it('returns defaults when query is empty', async () => {
    const { useSongFilters } = await import('./useSongFilters');
    const { filters } = useSongFilters();

    expect(filters.value.sort).toBe('created_at');
    expect(filters.value.order).toBe('desc');
    expect(filters.value.page).toBe(1);
    expect(filters.value.limit).toBe(20);
    expect(filters.value.q).toBeUndefined();
  });

  it('parses valid query params', async () => {
    currentQuery.value = {
      q: 'rock',
      artist: 'Beatles',
      sort: 'title',
      order: 'asc',
      page: '3',
      limit: '50',
    };

    const { useSongFilters } = await import('./useSongFilters');
    const { filters } = useSongFilters();

    expect(filters.value.q).toBe('rock');
    expect(filters.value.artist).toBe('Beatles');
    expect(filters.value.sort).toBe('title');
    expect(filters.value.order).toBe('asc');
    expect(filters.value.page).toBe(3);
    expect(filters.value.limit).toBe(50);
  });

  it('falls back to defaults for invalid sort/order', async () => {
    currentQuery.value = { sort: 'evil', order: 'sideways' };

    const { useSongFilters } = await import('./useSongFilters');
    const { filters } = useSongFilters();

    expect(filters.value.sort).toBe('created_at');
    expect(filters.value.order).toBe('desc');
  });

  it('falls back to default page when invalid', async () => {
    currentQuery.value = { page: '-1' };

    const { useSongFilters } = await import('./useSongFilters');
    const { filters } = useSongFilters();

    expect(filters.value.page).toBe(1);
  });

  it('update() resets page to 1 when page is not in patch', async () => {
    const { useSongFilters } = await import('./useSongFilters');
    const { update } = useSongFilters();

    update({ q: 'pop' });

    expect(pushSpy).toHaveBeenCalledTimes(1);
    const arg = pushSpy.mock.calls[0]![0];
    expect(arg.query.q).toBe('pop');
    expect(arg.query.page).toBe('1');
  });

  it('update() preserves explicit page when present in patch', async () => {
    const { useSongFilters } = await import('./useSongFilters');
    const { update } = useSongFilters();

    update({ page: 5 });

    const arg = pushSpy.mock.calls[0]![0];
    expect(arg.query.page).toBe('5');
  });

  it('update() with empty string removes the key', async () => {
    currentQuery.value = { q: 'previous' };

    const { useSongFilters } = await import('./useSongFilters');
    const { update } = useSongFilters();

    update({ q: '' });

    const arg = pushSpy.mock.calls[0]![0];
    expect(arg.query.q).toBeUndefined();
  });

  it('clear() pushes empty query', async () => {
    currentQuery.value = { q: 'rock', sort: 'title' };

    const { useSongFilters } = await import('./useSongFilters');
    const { clear } = useSongFilters();

    clear();

    const arg = pushSpy.mock.calls[0]![0];
    expect(arg.query).toEqual({});
  });
});
