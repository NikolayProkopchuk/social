package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/NikolayProkopchuk/social/internal/db"
	"github.com/NikolayProkopchuk/social/internal/env"
	"github.com/NikolayProkopchuk/social/internal/store"
)

func main() {
	cfg := config{
		address: env.GetString("ADDR", ":8080"),
		db: &DbConfig{
			url: fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
				env.GetString("DB_USER", "admin"),
				env.GetString("DB_PASSWORD", "admin_pwd"),
				env.GetString("DB_HOST", "localhost"),
				env.GetInt("DB_PORT", 5432),
				env.GetString("DB_NAME", "social")),
			maxOpenCons: env.GetInt("DB_MAX_OPEN_CONS", 10),
			maxIdleCons: env.GetInt("DB_MAX_IDLE_CONS", 10),
			maxIdleTime: env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env: env.GetString("ENV", "dev"),
	}
	d, err := db.New(
		cfg.db.url,
		cfg.db.maxOpenCons,
		cfg.db.maxIdleCons,
		cfg.db.maxIdleTime,
	)
	defer func(d *sql.DB) {
		err := d.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(d)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("DB connected")

	a := application{
		config: cfg,
		store:  store.NewStorage(d),
	}
	mux := a.mount()
	log.Fatal(a.run(mux))
}
