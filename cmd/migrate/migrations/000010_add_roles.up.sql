CREATE TABLE IF NOT EXISTS roles (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,
    level INT NOT NULL UNIQUE
);

INSERT INTO roles (name, description, level) VALUES
    ('user', 'Can create posts or comments', 1),
    ('moderator', 'Update other users posts', 2),
    ('admin', 'Update or delete other users posts', 3)
ON CONFLICT (name) DO NOTHING;
