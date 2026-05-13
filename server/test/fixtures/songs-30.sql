-- songs-30.sql
-- 30 songs owned by the regular test user with varied titles, artists,
-- and timestamps so filter/sort/pagination tests have meaningful data.

INSERT INTO songs (id, title, artist, duration, source_id, uploaded_by, created_at)
SELECT
  gen_random_uuid(),
  'Song ' || lpad(i::text, 2, '0'),
  CASE (i % 3) WHEN 0 THEN 'Artist A' WHEN 1 THEN 'Artist B' ELSE 'Artist C' END,
  180000 + (i * 1000),
  'src-' || i,
  '11111111-1111-1111-1111-111111111111',
  NOW() - (i * INTERVAL '1 hour')
FROM generate_series(1, 30) AS i;
