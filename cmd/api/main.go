package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/NikolayProkopchuk/social/internal/auth"
	"github.com/NikolayProkopchuk/social/internal/db"
	"github.com/NikolayProkopchuk/social/internal/env"
	"github.com/NikolayProkopchuk/social/internal/mailer"
	"github.com/NikolayProkopchuk/social/internal/store"
	"go.uber.org/zap"
)

//	@title			Gopher Social API
//	@description	This is a sample server Gopher Social.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/v1
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description				API Key for authorization
func main() {
	cfg := config{
		address: env.GetString("ADDR", ":8080"),
		apiUrl:  env.GetString("API_URL", "localhost:8080"),
		db: &dbConfig{
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
		mail: &mailConfig{
			exp:       24 * time.Hour,
			fromEmail: env.GetString("FROM_EMAIL", ""),
			sendgrid: sendgridConfig{
				apiKey: env.GetString("API_KEY", ""),
			},
		},
		frontednURL: env.GetString("FRONTEND_URL", "http://localhost:4000"),
		auth: &authConfig{
			basic: basicAuth{
				username: env.GetString("BASIC_AUTH_USERNAME", "admin"),
				password: env.GetString("BASIC_AUTH_PASSWORD", "password"),
			},
			tokenCfg: tokenConfig{
				secret:   env.GetString("AUTH_TOKEN_SECRET", "example"),
				issuer:   env.GetString("AUTH_TOKEN_ISSUER", "gopher.social"),
				audience: env.GetString("AUTH_TOKEN_AUDIENCE", "gopher.social"),
				exp:      time.Hour,
			},
		},
	}

	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	d, err := db.New(
		cfg.db.url,
		cfg.db.maxOpenCons,
		cfg.db.maxIdleCons,
		cfg.db.maxIdleTime,
	)
	defer func(d *sql.DB) {
		err := d.Close()
		if err != nil {
			logger.Fatal(err)
		}
	}(d)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("DB connected")

	mailerClient := mailer.NewSendGridMailer(cfg.mail.fromEmail, cfg.mail.sendgrid.apiKey)
	authenticator := auth.NewJWTAuthenticator(cfg.auth.tokenCfg.secret, cfg.auth.tokenCfg.issuer, cfg.auth.tokenCfg.issuer)

	a := application{
		config:        cfg,
		store:         store.NewStorage(d),
		logger:        logger,
		mailer:        mailerClient,
		authenticator: authenticator,
	}
	mux := a.mount()
	logger.Fatal(a.run(mux))
}
