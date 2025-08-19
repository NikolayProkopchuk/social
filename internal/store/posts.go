package store

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type PostStore struct {
	db *sql.DB
}

type Post struct {
	ID        int64      `json:"id"`
	Content   string     `json:"content"`
	Title     string     `json:"title"`
	UserID    int64      `json:"user_id"`
	Tags      []string   `json:"tags"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Version   int64      `json:"version"`
	Comments  []*Comment `json:"comments"`
}

func (s *PostStore) Create(ctx context.Context, post *Post) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimoutDuration)
	defer cancel()
	query := `
INSERT INTO posts (content, title, user_id, tags) VALUES ($1, $2, $3, $4)
RETURNING id, created_at, updated_at`
	return s.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserID,
		post.Tags).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt)
}

func (s *PostStore) Update(ctx context.Context, post *Post) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimoutDuration)
	defer cancel()
	query := `
UPDATE posts SET title = $1,
                 content = $2,
                 tags = $3,
                 updated_at = $4,
                 version = version + 1
             WHERE id = $5
               AND version = $6
             RETURNING version`
	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		post.Tags,
		time.Now(),
		post.ID,
		post.Version).Scan(&post.Version)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (s *PostStore) GetByID(ctx context.Context, id int64) (*Post, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimoutDuration)
	defer cancel()

	query := `
SELECT p.id, p.content, p.title, p.user_id, p.tags, p.created_at, p.updated_at, p.version FROM posts p WHERE p.id = $1`

	post := &Post{}
	m := pgtype.NewMap()
	err := s.db.QueryRowContext(
		ctx,
		query,
		id).Scan(
		&post.ID,
		&post.Content,
		&post.Title,
		&post.UserID,
		m.SQLScanner(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return post, nil
}

func (s *PostStore) DeleteByID(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimoutDuration)
	defer cancel()
	query := `
DELETE FROM posts WHERE id = $1`
	result, err := s.db.ExecContext(
		ctx,
		query,
		id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}

type PostWithMetadata struct {
	ID            int64     `json:"id"`
	Content       string    `json:"content"`
	Title         string    `json:"title"`
	UserID        int       `json:"userId"`
	Tags          []string  `json:"tags"`
	CreatedAt     time.Time `json:"createdAt"`
	CommentsCount int       `json:"commentsCount"`
}

func (s *PostStore) GetUserFeed(ctx context.Context, user *User, paginatedQuery *PaginatedFeedQuery) ([]*PostWithMetadata, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimoutDuration)
	defer cancel()
	query := `
SELECT p.id, p.content, p.title, p.user_id, p.tags, p.created_at, count(c.id) as comments_count
FROM posts p
LEFT JOIN comments c ON p.id = c.post_id
LEFT JOIN user_follower uf ON p.user_id = uf.follower_id
WHERE uf.user_id = $1
AND (p.title ILIKE '%' || $2 || '%' OR p.content ILIKE '%' || $2 || '%')
AND (p.tags @> $3 OR $3 IS NULL)
GROUP BY p.id, p.created_at
ORDER BY p.created_at ` + paginatedQuery.Sort +
		` LIMIT $4 OFFSET $5`
	log.Print(query)
	log.Print(paginatedQuery)
	rows, err := s.db.QueryContext(
		ctx,
		query,
		user.ID,
		paginatedQuery.Search,
		paginatedQuery.Tags,
		paginatedQuery.Limit,
		paginatedQuery.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	userFeed := make([]*PostWithMetadata, 0, paginatedQuery.Limit)
	m := pgtype.NewMap()
	for rows.Next() {
		post := &PostWithMetadata{}

		err = rows.Scan(
			&post.ID,
			&post.Content,
			&post.Title,
			&post.UserID,
			m.SQLScanner(&post.Tags),
			&post.CreatedAt,
			&post.CommentsCount)
		if err != nil {
			return nil, err
		}
		userFeed = append(userFeed, post)
	}
	return userFeed, nil
}
