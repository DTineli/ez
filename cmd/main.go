package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/DTineli/ez/internal/config"
	m "github.com/DTineli/ez/internal/middleware"
	"github.com/DTineli/ez/internal/store/dbstore"

	database "github.com/DTineli/ez/internal/store/db"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var Environment = "development"

func init() {
	os.Setenv("env", Environment)
	// run generate script
	exec.Command("make", "tailwind-build").Run()
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	r := chi.NewRouter()
	cfg := config.MustLoadConfig()

	r.Use(middleware.Logger)

	db := database.MustOpen(cfg.DatabaseName)
	sessionStore := dbstore.NewSessionStore(
		dbstore.NewSessionStoreParams{
			DB: db,
		},
	)

	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	authMiddleware := m.NewAuthMiddleware(sessionStore, cfg.SessionCookieName)

	r.Group(func(r chi.Router) {
		r.Use(
			middleware.Logger,
			m.TextHTMLMiddleware,
			authMiddleware.AddUserToContext,
		)
	})

	// r.Get("/login", handlers.NewGetLoginHandler().ServeHTTP)
	//
	// r.Post("/login", handlers.NewPostLoginHandler(handlers.PostLoginHandlerParams{
	// 	UserStore:         userStore,
	// 	SessionStore:      sessionStore,
	// 	PasswordHash:      passwordhash,
	// 	SessionCookieName: cfg.SessionCookieName,
	// }).ServeHTTP)
	//
	// r.Post("/logout", handlers.NewPostLogoutHandler(handlers.PostLogoutHandlerParams{
	// 	SessionCookieName: cfg.SessionCookieName,
	// }).ServeHTTP)

	killSig := make(chan os.Signal, 1)

	signal.Notify(killSig, os.Interrupt, syscall.SIGTERM)

	srv := &http.Server{
		Addr:    cfg.Port,
		Handler: r,
	}

	go func() {
		err := srv.ListenAndServe()

		if errors.Is(err, http.ErrServerClosed) {
			logger.Info("Server shutdown complete")
		} else if err != nil {
			logger.Error("Server error", slog.Any("err", err))
			os.Exit(1)
		}
	}()

	logger.Info("Server started", slog.String("port", cfg.Port), slog.String("env", Environment))
	<-killSig

	logger.Info("Shutting down server")

	// Create a context with a timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to gracefully shut down the server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown failed", slog.Any("err", err))
		os.Exit(1)
	}

	logger.Info("Server shutdown complete")
}
