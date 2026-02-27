-- +goose Up
CREATE TABLE IF NOT EXISTS roles (
    id   SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

INSERT INTO roles (name) VALUES ('admin'), ('user');

CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

ALTER TABLE songs ADD COLUMN IF NOT EXISTS uploaded_by UUID NULL REFERENCES users(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE songs DROP COLUMN IF EXISTS uploaded_by;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS roles;
