package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/NikolayProkopchuk/social/docs" // This line is used by Swag CLI to generate docs
	"github.com/NikolayProkopchuk/social/internal/mailer"
	"github.com/NikolayProkopchuk/social/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

var version = "0.0.1"

type application struct {
	config config
	store  *store.Storage
	logger *zap.SugaredLogger
	mailer mailer.Client
}

type config struct {
	address string
	db      *dbConfig
	env     string
	apiUrl  string
	mail    *mailConfig
	frontednURL string
}

type dbConfig struct {
	url         string
	maxOpenCons int
	maxIdleCons int
	maxIdleTime string
}

type mailConfig struct {
	sendgrid  sendgridConfig
	fromEmail string
	exp time.Duration
}

type sendgridConfig struct {
	apiKey    string
}

func (app *application) mount() http.Handler {
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiUrl
	docs.SwaggerInfo.BasePath = "/v1"

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)

		docsUrl := fmt.Sprintf("%s/swagger/doc.json", app.config.address)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsUrl)))
		r.Route("/posts", func(r chi.Router) {
			r.Post("/", app.createPostHandler)
			r.Route("/{postID}", func(r chi.Router) {
				r.Use(app.postsContextMiddleware)
				r.Get("/", app.getPostHandler)
				r.Patch("/", app.updatePostHandler)
				r.Delete("/", app.deletePostHandler)

				r.Route("/comments", func(r chi.Router) {
					r.Post("/", app.createCommentHandler)
				})
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Put("/active", app.activateUserHandler)
			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.userContextMiddleware)
				r.Get("/", app.getUserHandler)

				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})
			r.Group(func(r chi.Router) {
				r.Get("/feed", app.getUserFeedHandler)
			})
		})

		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
		})
	})

	return r
}

func (app *application) run(mux http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.address,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}
	app.logger.Infow("Starting server on", "addr:", app.config.address, "env:", app.config.env)
	return srv.ListenAndServe()
}
