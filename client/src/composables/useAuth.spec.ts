import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';

// useAuth uses module-level singleton state (see useAuth.ts:45-53). Each test
// resets modules so state doesn't bleed between cases.

describe('useAuth', () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    vi.resetModules();
    fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.useRealTimers();
  });

  it('login sets token + user and reports authenticated', async () => {
    fetchMock
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ access_token: 'tok-1', expires_in: 900 }),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({
          id: 'u-1',
          email: 'a@b.c',
          roles: ['user'],
          created_at: '2026-05-13T00:00:00Z',
          avatar_url: null,
        }),
      });

    const { useAuth } = await import('./useAuth');
    const auth = useAuth();
    const ok = await auth.login('a@b.c', 'secret-pw');

    expect(ok).toBe(true);
    expect(auth.isAuthenticated.value).toBe(true);
    expect(auth.user.value?.email).toBe('a@b.c');
    expect(auth.getAccessToken()).toBe('tok-1');
  });

  it('login failure surfaces error and stays unauthenticated', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      status: 401,
      json: async () => ({ error: 'bad creds' }),
    });

    const { useAuth } = await import('./useAuth');
    const auth = useAuth();
    const ok = await auth.login('a@b.c', 'wrong');

    expect(ok).toBe(false);
    expect(auth.error.value).toBe('bad creds');
    expect(auth.isAuthenticated.value).toBe(false);
  });

  it('fetchUser triggers refresh on 401, retries with new token, succeeds', async () => {
    fetchMock
      // login OK
      .mockResolvedValueOnce({
        ok: true, status: 200,
        json: async () => ({ access_token: 'tok-old', expires_in: 900 }),
      })
      // /me with old token → 401
      .mockResolvedValueOnce({
        ok: false, status: 401,
        json: async () => ({ error: 'expired' }),
      })
      // refresh → new token
      .mockResolvedValueOnce({
        ok: true, status: 200,
        json: async () => ({ access_token: 'tok-new', expires_in: 900 }),
      })
      // /me retried with new token → ok
      .mockResolvedValueOnce({
        ok: true, status: 200,
        json: async () => ({
          id: 'u-1', email: 'a@b.c', roles: ['user'],
          created_at: '2026-05-13T00:00:00Z', avatar_url: null,
        }),
      });

    const { useAuth } = await import('./useAuth');
    const auth = useAuth();
    await auth.login('a@b.c', 'pw');

    expect(auth.user.value?.email).toBe('a@b.c');
    expect(auth.getAccessToken()).toBe('tok-new');
  });

  it('refresh failure during fetchUser clears auth state', async () => {
    fetchMock
      .mockResolvedValueOnce({
        ok: true, status: 200,
        json: async () => ({ access_token: 'tok-old', expires_in: 900 }),
      })
      .mockResolvedValueOnce({
        ok: false, status: 401, json: async () => ({}),
      })
      .mockResolvedValueOnce({
        ok: false, status: 401, json: async () => ({}),
      });

    const { useAuth } = await import('./useAuth');
    const auth = useAuth();
    await auth.login('a@b.c', 'pw');

    expect(auth.isAuthenticated.value).toBe(false);
    expect(auth.user.value).toBeNull();
  });

  it('logout clears state even when network call fails', async () => {
    fetchMock
      .mockResolvedValueOnce({
        ok: true, status: 200,
        json: async () => ({ access_token: 'tok-1', expires_in: 900 }),
      })
      .mockResolvedValueOnce({
        ok: true, status: 200,
        json: async () => ({
          id: 'u-1', email: 'a@b.c', roles: ['user'],
          created_at: '2026-05-13T00:00:00Z', avatar_url: null,
        }),
      });

    const { useAuth } = await import('./useAuth');
    const auth = useAuth();
    await auth.login('a@b.c', 'pw');

    // Now logout — network fails
    fetchMock.mockRejectedValueOnce(new Error('boom'));
    await auth.logout();

    expect(auth.isAuthenticated.value).toBe(false);
    expect(auth.user.value).toBeNull();
  });

  it('isAdmin is true when roles include "admin"', async () => {
    fetchMock
      .mockResolvedValueOnce({
        ok: true, status: 200,
        json: async () => ({ access_token: 'tok-1', expires_in: 900 }),
      })
      .mockResolvedValueOnce({
        ok: true, status: 200,
        json: async () => ({
          id: 'u-1', email: 'a@b.c', roles: ['user', 'admin'],
          created_at: '2026-05-13T00:00:00Z', avatar_url: null,
        }),
      });

    const { useAuth } = await import('./useAuth');
    const auth = useAuth();
    await auth.login('a@b.c', 'pw');

    expect(auth.isAdmin.value).toBe(true);
  });
});
