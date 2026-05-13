import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';

// useAvatarUpload depends on useApi which depends on useAuth (module-level
// singletons). We reset modules and mock useApi to avoid cross-test bleed.

describe('useAvatarUpload', () => {
  beforeEach(() => {
    vi.resetModules();
    // happy-dom does not implement createObjectURL — stub it.
    if (typeof URL.createObjectURL !== 'function') {
      (URL as unknown as { createObjectURL: (f: Blob) => string }).createObjectURL =
        () => 'blob:fake-preview';
      (URL as unknown as { revokeObjectURL: (s: string) => void }).revokeObjectURL =
        () => {};
    }
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('rejects file with unsupported extension and surfaces error', async () => {
    vi.doMock('./useApi', () => ({
      useApi: () => ({
        uploadFile: vi.fn(),
        setOwnAvatar: vi.fn(),
        clearOwnAvatar: vi.fn(),
        setUserAvatar: vi.fn(),
        clearUserAvatar: vi.fn(),
      }),
    }));

    const { useAvatarUpload } = await import('./useAvatarUpload');
    const upload = useAvatarUpload({ target: 'self' });

    const txt = new File(['hello'], 'evil.txt', { type: 'text/plain' });
    await upload.pick(txt);

    expect(upload.state.value).toBe('error');
    expect(upload.error.value).toContain('Unsupported format');
  });

  it('rejects file over MAX_AVATAR_BYTES', async () => {
    const { MAX_AVATAR_BYTES } = await import('../types/api');

    vi.doMock('./useApi', () => ({
      useApi: () => ({
        uploadFile: vi.fn(),
      }),
    }));

    const { useAvatarUpload } = await import('./useAvatarUpload');
    const upload = useAvatarUpload({ target: 'self' });

    const bytes = new Uint8Array(MAX_AVATAR_BYTES + 1);
    const big = new File([bytes], 'huge.jpg', { type: 'image/jpeg' });
    await upload.pick(big);

    expect(upload.state.value).toBe('error');
    expect(upload.error.value).toContain('too large');
  });

  it('uploads valid image, transitions to previewing', async () => {
    const uploadFile = vi.fn().mockResolvedValue({
      data: { hash: 'abc123', content_type: 'image/jpeg', size_bytes: 100 },
      error: null,
      status: 201,
    });

    vi.doMock('./useApi', () => ({
      useApi: () => ({ uploadFile, setOwnAvatar: vi.fn() }),
    }));

    const { useAvatarUpload } = await import('./useAvatarUpload');
    const upload = useAvatarUpload({ target: 'self' });

    const img = new File([new Uint8Array([0xff, 0xd8])], 'me.jpg', { type: 'image/jpeg' });
    await upload.pick(img);

    expect(upload.state.value).toBe('previewing');
    expect(uploadFile).toHaveBeenCalledWith(img);
    // happy-dom generates its own blob URL — assert shape, not literal value.
    expect(upload.previewUrl.value).toMatch(/^blob:/);
  });

  it('confirm() calls setOwnAvatar and transitions to committed', async () => {
    const uploadFile = vi.fn().mockResolvedValue({
      data: { hash: 'abc123', content_type: 'image/jpeg', size_bytes: 100 },
      error: null,
      status: 201,
    });
    const setOwnAvatar = vi.fn().mockResolvedValue({
      data: { avatar_url: '/api/files/abc123' },
      error: null,
      status: 200,
    });

    vi.doMock('./useApi', () => ({
      useApi: () => ({ uploadFile, setOwnAvatar }),
    }));

    const { useAvatarUpload } = await import('./useAvatarUpload');
    const upload = useAvatarUpload({ target: 'self' });

    const img = new File([new Uint8Array([0xff, 0xd8])], 'me.jpg', { type: 'image/jpeg' });
    await upload.pick(img);
    const ok = await upload.confirm();

    expect(ok).toBe(true);
    expect(setOwnAvatar).toHaveBeenCalledWith('abc123');
    expect(upload.state.value).toBe('committed');
    expect(upload.committedUrl.value).toBe('/api/files/abc123');
  });

  it('cancel() returns to idle state and clears preview', async () => {
    const uploadFile = vi.fn().mockResolvedValue({
      data: { hash: 'h', content_type: 'image/jpeg', size_bytes: 1 },
      error: null,
      status: 201,
    });

    vi.doMock('./useApi', () => ({
      useApi: () => ({ uploadFile, setOwnAvatar: vi.fn() }),
    }));

    const { useAvatarUpload } = await import('./useAvatarUpload');
    const upload = useAvatarUpload({ target: 'self' });

    const img = new File([new Uint8Array([0xff, 0xd8])], 'me.jpg', { type: 'image/jpeg' });
    await upload.pick(img);
    expect(upload.state.value).toBe('previewing');

    upload.cancel();

    expect(upload.state.value).toBe('idle');
    expect(upload.previewUrl.value).toBeNull();
  });
});
