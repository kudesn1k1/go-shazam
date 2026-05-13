-- users-roles.sql
-- Seeds two known users with deterministic UUIDs for integration & E2E tests.
-- The bcrypt hash below is for the password "test-password-123" (cost 10).
-- Roles 'user' and 'admin' are already inserted by migration 20241130000000_roles.sql.

INSERT INTO users (id, email, email_hash, hashed_password, created_at, updated_at)
VALUES
  ('11111111-1111-1111-1111-111111111111',
   'encrypted-user-email',
   'user-email-hash',
   '$2a$10$ZXqcuOuI8N4uZcfvxK1ZyOq8xL2y/wU9G9ZQ8WqXk1bTQ5y0eXfTu',
   NOW(), NOW()),
  ('22222222-2222-2222-2222-222222222222',
   'encrypted-admin-email',
   'admin-email-hash',
   '$2a$10$ZXqcuOuI8N4uZcfvxK1ZyOq8xL2y/wU9G9ZQ8WqXk1bTQ5y0eXfTu',
   NOW(), NOW())
ON CONFLICT (email_hash) DO NOTHING;

INSERT INTO user_roles (user_id, role_id)
SELECT '11111111-1111-1111-1111-111111111111', id FROM roles WHERE name = 'user'
ON CONFLICT DO NOTHING;

INSERT INTO user_roles (user_id, role_id)
SELECT '22222222-2222-2222-2222-222222222222', id FROM roles WHERE name = 'user'
ON CONFLICT DO NOTHING;

INSERT INTO user_roles (user_id, role_id)
SELECT '22222222-2222-2222-2222-222222222222', id FROM roles WHERE name = 'admin'
ON CONFLICT DO NOTHING;
