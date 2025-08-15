package store

import (
	"context"
	"database/sql"
	"time"
)

type Comment struct {
	ID        int64        `json:"id"`
	PostID    int64        `json:"post_id"`
	Content   string       `json:"content"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt sql.NullTime `json:"updated_at"`
	User      *User        `json:"user"`
}

type CommentStore struct {
	db *sql.DB
}

func (s *CommentStore) GetByPostID(ctx context.Context, postID int64) ([]*Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimoutDuration)
	defer cancel()
	query := `
SELECT c.id,
       c.post_id,
       c.user_id,
       u.username,
       c.content,
       c.created_at,
       c.updated_at
FROM comments c
JOIN users u ON u.id = c.user_id
WHERE c.post_id = $1`

	rows, err := s.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []*Comment
	for rows.Next() {
		var comment Comment
		var user User
		comment.User = &user
		if err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.User.ID,
			&comment.User.Username,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		); err != nil {
			return nil, err
		}
		comments = append(comments, &comment)
	}

	return comments, nil
}

func (s *CommentStore) Create(ctx context.Context, comment *Comment) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimoutDuration)
	defer cancel()
	query := `
INSERT INTO comments (post_id, user_id, content) VALUES ($1, $2, $3)
RETURNING id, created_at, updated_at`
	return s.db.QueryRowContext(
		ctx,
		query,
		comment.PostID,
		comment.User.ID,
		comment.Content,
	).Scan(
		&comment.ID,
		&comment.CreatedAt,
		&comment.UpdatedAt)
}
