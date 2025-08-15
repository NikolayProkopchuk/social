package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

type FollowerStore struct {
	db *sql.DB
}

func (s *FollowerStore) Follow(ctx context.Context, user *User, follower *User) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimoutDuration)
	defer cancel()
	query := `
INSERT INTO user_follower (user_id, follower_id) VALUES ($1, $2)`
	_, err := s.db.ExecContext(
		ctx,
		query,
		user.ID,
		follower.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrConflict
		}
		return err
	}
	return nil
}

func (s *FollowerStore) Unfollow(ctx context.Context, user *User, follower *User) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimoutDuration)
	defer cancel()
	query := `
DELETE FROM user_follower WHERE user_id = $1 AND follower_id = $2`
	_, err := s.db.ExecContext(
		ctx,
		query,
		user.ID,
		follower.ID)
	return err
}
