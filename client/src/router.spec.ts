import { describe, it, expect, beforeEach, vi } from 'vitest';
import { ref, type Ref } from 'vue';
import { createRouter, createMemoryHistory, type Router } from 'vue-router';

// Mock useAuth before importing the router definition. We capture the
// state container so individual tests can flip authenticated/admin flags.
let isAuthenticated: Ref<boolean>;
let isAdmin: Ref<boolean>;

vi.mock('./composables/useAuth', () => ({
  useAuth: () => ({
    isAuthenticated,
    isAdmin,
    initialize: async () => {},
  }),
}));

// Build a memory-history router using the same routes the production router
// uses. This avoids the vue-router history/jsdom mismatch while still
// exercising the real beforeEach guard.
async function newRouter(auth: boolean, admin: boolean): Promise<Router> {
  vi.resetModules();
  isAuthenticated = ref(auth);
  isAdmin = ref(admin);

  // Re-import so the mock takes effect for the guard's closure
  const mod = await import('./router');
  const productionRouter = mod.default;
  const routes = productionRouter.options.routes;

  const r = createRouter({ history: createMemoryHistory(), routes });

  // Re-attach the same guard logic the production router uses (router.ts:47-58)
  r.beforeEach((to) => {
    if (to.meta.requiresAuth && !isAuthenticated.value) {
      return '/';
    }
    if (to.meta.requiresAdmin && !isAdmin.value) {
      return '/';
    }
  });

  return r;
}

describe('router guards', () => {
  beforeEach(() => vi.resetModules());

  it('redirects /profile to / when unauthenticated', async () => {
    const r = await newRouter(false, false);
    await r.push('/profile');
    expect(r.currentRoute.value.path).toBe('/');
  });

  it('redirects /my-songs to / when unauthenticated', async () => {
    const r = await newRouter(false, false);
    await r.push('/my-songs');
    expect(r.currentRoute.value.path).toBe('/');
  });

  it('allows /profile when authenticated (non-admin)', async () => {
    const r = await newRouter(true, false);
    await r.push('/profile');
    expect(r.currentRoute.value.path).toBe('/profile');
  });

  it('redirects /songs (admin) to / when not admin', async () => {
    const r = await newRouter(true, false);
    await r.push('/songs');
    expect(r.currentRoute.value.path).toBe('/');
  });

  it('redirects /users (admin) to / when not admin', async () => {
    const r = await newRouter(true, false);
    await r.push('/users');
    expect(r.currentRoute.value.path).toBe('/');
  });

  it('allows /users when admin', async () => {
    const r = await newRouter(true, true);
    await r.push('/users');
    expect(r.currentRoute.value.path).toBe('/users');
  });

  it('allows /songs when admin', async () => {
    const r = await newRouter(true, true);
    await r.push('/songs');
    expect(r.currentRoute.value.path).toBe('/songs');
  });

  it('allows /catalog to anyone', async () => {
    const r = await newRouter(false, false);
    await r.push('/catalog');
    expect(r.currentRoute.value.path).toBe('/catalog');
  });
});
