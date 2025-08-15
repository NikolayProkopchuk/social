CREATE TABLE IF NOT EXISTS user_follower (
    user_id BIGINT NOT NULL,
    follower_id BIGINT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pk_user_follower PRIMARY KEY (user_id, follower_id)
);
