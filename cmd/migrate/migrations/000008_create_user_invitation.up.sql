CREATE TABLE IF NOT EXISTS user_invitation (
    user_id BIGINT CONSTRAINT pk_user_invitation PRIMARY KEY,
    CONSTRAINT fk_user_invitation_user_id FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE,
    invite_code bytea NOT NULL UNIQUE,
    expiration_time TIMESTAMP NOT NULL
);
