ALTER TABLE IF EXISTS comments
    DROP CONSTRAINT IF EXISTS fk_comments_post_id,
    ADD CONSTRAINT fk_comments_post_id FOREIGN KEY (post_id) REFERENCES posts (id) ON DELETE CASCADE,
    DROP CONSTRAINT IF EXISTS fk_comments_user_id,
    ADD CONSTRAINT fk_comments_user_id FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;
