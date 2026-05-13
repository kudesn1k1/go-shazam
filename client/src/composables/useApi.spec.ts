import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';

describe('useApi', () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    vi.resetModules();
    fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('returns 401 surface as error (no infinite refresh loop in this layer)', async () => {
    // useApi delegates token retrieval to useAuth.getAccessToken() but does NOT
    // implement its own 401-refresh-retry — that lives in useAuth.fetchUser.
    // So this test asserts that useApi simply surfaces the 401 error to the
    // caller. No retries, no loops.
    fetchMock.mockResolvedValue({
      ok: false,
      status: 401,
      json: async () => ({ error: 'expired' }),
    });

    const { useApi } = await import('./useApi');
    const api = useApi();

    const r1 = await api.listMySongs({ sort: 'created_at', order: 'desc', page: 1, limit: 20 });
    const r2 = await api.listMySongs({ sort: 'created_at', order: 'desc', page: 1, limit: 20 });

    expect(r1.status).toBe(401);
    expect(r1.error).toBe('expired');
    expect(r2.status).toBe(401);
    // Exactly one fetch per call, no implicit retry loop
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it('attaches Authorization header when token present', async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ data: [], total: 0, page: 1, limit: 20 }),
    });

    // Establish an auth token via the singleton state. We have to mock useAuth
    // for this to work because both modules share state.
    vi.doMock('./useAuth', () => ({
      useAuth: () => ({ getAccessToken: () => 'tok-test' }),
    }));

    const { useApi } = await import('./useApi');
    const api = useApi();

    await api.listAllSongs({ sort: 'created_at', order: 'desc', page: 1, limit: 20 });

    const [, init] = fetchMock.mock.calls[0]!;
    const headers = (init as RequestInit).headers as Record<string, string>;
    expect(headers.Authorization).toBe('Bearer tok-test');
  });

  it('uploadFile sends multipart body without Content-Type override', async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      status: 201,
      json: async () => ({ hash: 'h', content_type: 'image/jpeg', size_bytes: 1 }),
    });

    vi.doMock('./useAuth', () => ({
      useAuth: () => ({ getAccessToken: () => 'tok-test' }),
    }));

    const { useApi } = await import('./useApi');
    const api = useApi();

    const file = new File([new Uint8Array([0xff, 0xd8])], 'a.jpg', { type: 'image/jpeg' });
    await api.uploadFile(file);

    const [, init] = fetchMock.mock.calls[0]!;
    const headers = (init as RequestInit).headers as Record<string, string>;
    // Content-Type must NOT be set for FormData bodies — browser must set it
    // with the multipart boundary.
    expect(headers['Content-Type']).toBeUndefined();
    expect((init as RequestInit).body).toBeInstanceOf(FormData);
  });

  it('surfaces network errors as a normalized error', async () => {
    fetchMock.mockRejectedValue(new Error('network down'));

    const { useApi } = await import('./useApi');
    const api = useApi();

    const { error, status, data } = await api.listPublicSongs({
      sort: 'created_at',
      order: 'desc',
      page: 1,
      limit: 20,
    });

    expect(error).toContain('Network error');
    expect(status).toBe(0);
    expect(data).toBeNull();
  });
});
