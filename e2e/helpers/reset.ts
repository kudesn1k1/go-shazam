import { request } from '@playwright/test';

const BASE_URL = process.env.BASE_URL ?? 'http://localhost';

export async function resetServer(baseURL = BASE_URL): Promise<void> {
  const api = await request.newContext({ baseURL });
  const res = await api.post('/api/test/reset');
  if (!res.ok()) {
    throw new Error(`reset failed: ${res.status()} ${await res.text()}`);
  }
  await api.dispose();
}

export interface SeedSongInput {
  title: string;
  artist: string;
  duration?: number;
  source_id?: string;
  uploaded_by?: string;
}

export async function seedSong(
  input: SeedSongInput,
  baseURL = BASE_URL,
): Promise<string> {
  const api = await request.newContext({ baseURL });
  const res = await api.post('/api/test/seed-song', {
    data: {
      duration: 180000,
      source_id: `src-${Date.now()}-${Math.random()}`,
      ...input,
    },
  });
  if (!res.ok()) {
    throw new Error(`seed-song failed: ${res.status()} ${await res.text()}`);
  }
  const body = (await res.json()) as { id: string };
  await api.dispose();
  return body.id;
}

export async function promote(
  emailHash: string,
  role: 'admin' | 'user',
  baseURL = BASE_URL,
): Promise<void> {
  const api = await request.newContext({ baseURL });
  const res = await api.post('/api/test/promote', {
    data: { email_hash: emailHash, role },
  });
  if (!res.ok()) {
    throw new Error(`promote failed: ${res.status()} ${await res.text()}`);
  }
  await api.dispose();
}
