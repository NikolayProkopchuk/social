package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func New(url string, maxOpenCons int, maxIdleCons int, maxIdleTimeStr string) (*sql.DB, error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(maxOpenCons)
	db.SetMaxIdleConns(maxIdleCons)
	idleTime, err := time.ParseDuration(maxIdleTimeStr)
	db.SetConnMaxIdleTime(idleTime)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}
	return db, nil
}
