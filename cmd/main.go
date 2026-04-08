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
	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/store/cookiesotore"
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
	userStore := dbstore.NewUserStore(db)

	sessionStore := cookiesotore.NewSessionStore(
		store.AdminSessionName,
		"VERYSECRETKEY", // TODO: Colocar no env
	)

	clientSessionStore := cookiesotore.NewSessionStore(
		store.ClientSessionName,
		"VERYSECRETKEY", // TODO: Colocar no env
	)

	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	invite := dbstore.NewInvireStore(db)
	tenantStore := dbstore.NewTenantStore(db)
	contactStore := dbstore.NewContactStore(db)

	registerHandler := handlers.NewRegisterHandler(
		userStore,
		tenantStore,
		invite,
		contactStore,
	)

	loginHandler := handlers.NewLoginHandler(
		handlers.LoginHandlerParams{
			UserStore:    userStore,
			SessionStore: sessionStore,
			TenantStore:  *tenantStore,
			CookieName:   cfg.SessionCookieName,
		},
	)

	//Creating handlers
	productHandler := handlers.NewProductHandler(
		dbstore.NewProductStore(db),
		dbstore.NewPriceTableDB(db),
	)

	contactHandler := handlers.NewContactHandler(
		handlers.NewContactHandlerParams{
			Contact: contactStore,
			Invite:  invite,
		},
	)

	r.Route("/client", func(r chi.Router) {
		r.Use(m.TextHTMLMiddleware)

		r.Get("/login", loginHandler.GetClientLoginPage)
		r.Post("/login", loginHandler.PostLoginHandler(store.AccessCustomer))

		r.Get("/register", registerHandler.GetRegisterClientPage)
		r.Post("/register", registerHandler.PostRegisterClient)

		r.Group(func(r chi.Router) {
			r.Use(m.SessionAuthMiddleware(clientSessionStore))

			r.Post("/logout", loginHandler.PostLogout)
			r.Get("/", func(w http.ResponseWriter, req *http.Request) {
				w.Write([]byte("<h1>Vai Corinthians</h1>"))
			})
		})
	})

	r.Route("/admin", func(r chi.Router) {
		r.Use(m.TextHTMLMiddleware)

		r.Get("/", loginHandler.GetAdminLoginPage)

		r.Get("/login", loginHandler.GetAdminLoginPage)
		r.Post("/login", loginHandler.PostLoginHandler(store.AccessAdmin))
		r.Get("/register", registerHandler.GetRegisterPage)
		r.Post("/register", registerHandler.PostRegister)

		r.Post("/logout", loginHandler.PostLogout)

		r.Group(func(r chi.Router) {
			r.Use(
				m.SessionAuthMiddleware(sessionStore),
			)

			r.Get("/", handlers.NewHomeHandler(sessionStore).ServeHTTP)

			r.Route("/produtos", func(r chi.Router) {
				r.Get("/", productHandler.GetProductPage)
				r.Get("/novo", productHandler.GetProductForm)
				r.Get("/{id}", productHandler.GetEditPage)

				r.Get("/pricetable", productHandler.GetTablePage)
				r.Post("/pricetable", productHandler.CreatePriceTable)

				r.Post("/", productHandler.PostNewProduct)
				r.Post("/{id}", productHandler.UpdateProduct)
				r.Delete("/{id}", productHandler.DeleteProduct)
			})

			r.Route("/contacts", func(r chi.Router) {
				r.Post("/", contactHandler.PostNewContact)
				r.Post("/{id}/create-link", contactHandler.CreateLink)

				r.Post("/{id}", contactHandler.Update)

				r.Get("/{id}", contactHandler.GetEditPage)

				r.Get("/", contactHandler.GetContactsPage)
				r.Get("/novo", contactHandler.GetContactsForm)
			})
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
