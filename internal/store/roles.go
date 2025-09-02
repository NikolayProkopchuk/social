package store

import (
	"context"
	"database/sql"
)

type RoleStore struct {
	db *sql.DB
}

type Role struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Level       int    `json:"level,omitempty"`
}

func (s *RoleStore) GetByName(ctx context.Context, name string) (*Role, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimoutDuration)
	defer cancel()
	query := `SELECT id, name, description, level FROM roles WHERE name = $1`
	var role Role
	err := s.db.QueryRowContext(
		ctx, query, name).Scan(
		&role.ID, &role.Name, &role.Description, &role.Level)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}
