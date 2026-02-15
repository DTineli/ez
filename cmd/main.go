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
	"github.com/DTineli/ez/internal/handlers"
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
	// r.Use(m.CheckTenantMiddleware)

	db := database.MustOpen(cfg.DatabaseName)
	userStore := dbstore.NewUserStore(db)

	sessionStore := dbstore.NewSessionStore(
		dbstore.NewSessionStoreParams{
			DB: db,
		},
	)

	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	authMiddleware := m.NewAuthMiddleware(sessionStore, cfg.SessionCookieName)

	tenantStore := dbstore.NewTenantStore(db)
	registerHandler := handlers.NewRegisterHandler(userStore, tenantStore)

	loginHandler := handlers.NewLoginHandler(userStore, sessionStore, cfg.SessionCookieName)

	r.Group(func(r chi.Router) {
		// r.Use(m.TextHTMLMiddleware)

		r.Get("/login", loginHandler.GetLoginPage)
		r.Post("/login", loginHandler.PostLogin)
		r.Get("/register", registerHandler.GetRegisterPage)
		r.Post("/register", registerHandler.PostRegister)

		r.Post("/logout", loginHandler.PostLogout)
	})

	productHandler := handlers.NewProductHandler(dbstore.NewProductStore(db))

	// autenticado
	r.Group(func(r chi.Router) {
		r.Use(m.TextHTMLMiddleware, authMiddleware.AddUserToContext)
		r.Get("/", handlers.NewHomeHandler().ServeHTTP)

		r.Route("/produtos", func(r chi.Router) {
			r.Get("/", productHandler.GetProductPage)
			r.Post("/", productHandler.PostNewProduct)
			r.Get("/novo", productHandler.GetProductForm)
			r.Delete("/{id}", productHandler.DeleteProduct)

			// r.Get("/{id}", produtoHandler.Show)
		})
	})

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
