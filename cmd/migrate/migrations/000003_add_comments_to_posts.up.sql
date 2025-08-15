CREATE TABLE IF NOT EXISTS comments (
    id BIGSERIAL CONSTRAINT pk_comments PRIMARY KEY,
    post_id BIGINT NOT NULL CONSTRAINT fk_comments_post_id REFERENCES posts(id),
    user_id BIGINT NOT NULL CONSTRAINT fk_comments_user_id REFERENCES users(id),
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP
);
