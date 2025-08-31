package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type password struct {
	text string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err == nil {
		p.hash = hash
		p.text = text
	}
	return err
}

func (p *password) Compare(text string) error {
	return bcrypt.CompareHashAndPassword(p.hash, []byte(text))
}

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) create(ctx context.Context, trx *sql.Tx, user *User) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimoutDuration)
	defer cancel()
	query := `
INSERT INTO users (username, email, password) VALUES ($1, $2, $3)
RETURNING id, created_at`
	err := trx.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password.hash).Scan(
		&user.ID,
		&user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrConflict
		}
		return err
	}
	return nil
}

func (s *UserStore) GetByID(ctx context.Context, id int64) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimoutDuration)
	defer cancel()
	query := `
SELECT id, username, email, created_at FROM users WHERE id = $1`
	user := &User{}
	err := s.db.QueryRowContext(
		ctx,
		query,
		id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return user, nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimoutDuration)
	defer cancel()
	query := `
SELECT id, username, email, password, created_at FROM users WHERE email = $1`
	user := &User{}
	err := s.db.QueryRowContext(
		ctx,
		query,
		email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return user, nil
}

func (s *UserStore) CreateAndInvite(ctx context.Context, user *User, inviteCode string, expirationTime time.Duration) error {
	return withTrx(ctx, s.db, func(trx *sql.Tx) error {
		if err := s.create(ctx, trx, user); err != nil {
			return err
		}
		if err := s.createUserInvitation(ctx, trx, user.ID, inviteCode, expirationTime); err != nil {
			return err
		}
		return nil
	})
}

func (s *UserStore) createUserInvitation(ctx context.Context, trx *sql.Tx, userID int64, inviteCode string, expirationTime time.Duration) error {
	query := `INSERT INTO user_invitation (user_id, invite_code, expiration_time) VALUES ($1, $2, $3)`
	ctx, cancel := context.WithTimeout(ctx, QueryTimoutDuration)
	defer cancel()
	_, err := trx.ExecContext(
		ctx,
		query,
		userID,
		inviteCode,
		time.Now().Add(expirationTime))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrConflict
		}
		return err
	}
	return nil
}

func (s *UserStore) Activate(ctx context.Context, inviteCodeHashed string) error {
	return withTrx(ctx, s.db, func(trx *sql.Tx) error {
		if err := s.activate(ctx, trx, inviteCodeHashed); err != nil {
			return err
		}
		if err := s.deleteInvite(ctx, trx, inviteCodeHashed); err != nil {
			return err
		}
		return nil
	})
}

func (s *UserStore) activate(ctx context.Context, tx *sql.Tx, inviteCodeHashed string) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimoutDuration)
	defer cancel()
	query := `
UPDATE users SET active = TRUE
WHERE id = (
	SELECT user_id FROM user_invitation
	WHERE invite_code = $1 AND expiration_time > NOW()
)`
	res, err := tx.ExecContext(ctx, query, inviteCodeHashed)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *UserStore) deleteInvite(ctx context.Context, tx *sql.Tx, inviteCodeHashed string) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimoutDuration)
	defer cancel()
	query := `DELETE FROM user_invitation WHERE invite_code = $1`
	_, err := tx.ExecContext(ctx, query, inviteCodeHashed)
	if err != nil {
		return err
	}
	return nil
}
