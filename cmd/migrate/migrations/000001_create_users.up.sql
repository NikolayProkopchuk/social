CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users
(
    id         BIGSERIAL PRIMARY KEY,
    username   TEXT                        NOT NULL UNIQUE,
    email      citext UNIQUE               NOT NULL,
    password   bytea                       NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
