import pg from 'pg';

const PG_CONFIG = {
  host: process.env.DB_HOST ?? 'localhost',
  port: Number(process.env.DB_PORT ?? 5432),
  user: process.env.DB_USER ?? 'go-shazam-db-user',
  password: process.env.DB_PASSWORD ?? 'go-shazam-db-password',
  database: process.env.DB_DATABASE ?? 'go-shazam-database',
};

export async function withDb<T>(fn: (c: pg.Client) => Promise<T>): Promise<T> {
  const client = new pg.Client(PG_CONFIG);
  await client.connect();
  try {
    return await fn(client);
  } finally {
    await client.end();
  }
}

export async function countSongs(): Promise<number> {
  return withDb(async (c) => {
    const r = await c.query<{ n: string }>('SELECT COUNT(*)::text AS n FROM songs');
    return Number(r.rows[0].n);
  });
}

export async function getEmailHashByUserId(userId: string): Promise<string> {
  return withDb(async (c) => {
    const r = await c.query<{ email_hash: string }>(
      'SELECT email_hash FROM users WHERE id = $1',
      [userId],
    );
    if (r.rows.length === 0) throw new Error(`no user with id ${userId}`);
    return r.rows[0].email_hash;
  });
}

export async function getLatestEmailHash(): Promise<string> {
  return withDb(async (c) => {
    const r = await c.query<{ email_hash: string }>(
      'SELECT email_hash FROM users ORDER BY created_at DESC LIMIT 1',
    );
    if (r.rows.length === 0) throw new Error('no users in DB');
    return r.rows[0].email_hash;
  });
}
