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
	if Environment == "development" {
		exec.Command("make", "tailwind-build").Run()
	}
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.MustLoadConfig()

	db := database.MustOpen(cfg.DatabaseName)

	// stores
	userStore := dbstore.NewUserStore(db)
	tenantStore := dbstore.NewTenantStore(db)
	invite := dbstore.NewInvireStore(db)
	contactStore := dbstore.NewContactStore(db)
	pStore := dbstore.NewProductStore(db)
	priceTableStore := dbstore.NewPriceTableDB(db)
	cartStore := dbstore.NewCartStore(db)
	orderStore := dbstore.NewOrderStore(db)

	// sessions
	sessionStore := cookiesotore.NewSessionStore(store.AdminSessionName, cfg.SessionSecret)
	clientSessionStore := cookiesotore.NewSessionStore(store.ClientSessionName, cfg.SessionSecret)

	// handlers
	loginHandler := handlers.NewLoginHandler(handlers.LoginHandlerParams{
		UserStore:    userStore,
		SessionStore: sessionStore,
		TenantStore:  *tenantStore,
		CookieName:   cfg.SessionCookieName,
	})
	registerHandler := handlers.NewRegisterHandler(userStore, tenantStore, invite, contactStore, clientSessionStore)
	productHandler := handlers.NewProductHandler(pStore, priceTableStore)
	contactHandler := handlers.NewContactHandler(handlers.NewContactHandlerParams{
		Contact:    contactStore,
		Invite:     invite,
		PriceTable: priceTableStore,
	})
	clientHandler := handlers.NewClientHandler(pStore, cartStore, orderStore, clientSessionStore, priceTableStore)
	adminOrderHandler := handlers.NewAdminOrderHandler(orderStore, contactStore, pStore)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	registerClientRoutes(r, loginHandler, registerHandler, clientHandler, clientSessionStore)
	registerAdminRoutes(r, loginHandler, registerHandler, productHandler, contactHandler, adminOrderHandler, sessionStore)

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown failed", slog.Any("err", err))
		os.Exit(1)
	}

	logger.Info("Server shutdown complete")
}

func registerClientRoutes(
	r chi.Router,
	login *handlers.LoginHandler,
	reg *handlers.RegisterHandler,
	client *handlers.ClientHandler,
	sessionStore *cookiesotore.SessionStore,
) {
	r.Route("/client", func(r chi.Router) {
		r.Use(m.TextHTMLMiddleware)

		r.Get("/login", login.GetClientLoginPage)
		r.Post("/login", login.PostLoginHandler(store.AccessCustomer))
		r.Get("/register", reg.GetRegisterClientPage)
		r.Post("/register", reg.PostRegisterClient)

		r.Group(func(r chi.Router) {
			r.Use(m.SessionAuthMiddleware(sessionStore))

			r.Post("/logout", login.PostLogout)
			r.Get("/items", client.GetItemsPage)
			r.Get("/confirmacao", client.GetCheckoutPage)
			r.Post("/cart/items", client.PostAddToCart)
			r.Delete("/cart/items/{productID}", client.DeleteCartItem)
			r.Patch("/cart/items/{productID}", client.PatchCartItemQty)
			r.Post("/confirmacao", client.PostConfirmOrder)
		})
	})
}

func registerAdminRoutes(
	r chi.Router,
	login *handlers.LoginHandler,
	reg *handlers.RegisterHandler,
	product *handlers.ProductHandler,
	contact *handlers.ContactHandler,
	order *handlers.AdminOrderHandler,
	sessionStore *cookiesotore.SessionStore,
) {
	r.Route("/admin", func(r chi.Router) {
		r.Use(m.TextHTMLMiddleware)

		r.Get("/", login.GetAdminLoginPage)
		r.Get("/login", login.GetAdminLoginPage)
		r.Post("/login", login.PostLoginHandler(store.AccessAdmin))
		r.Get("/register", reg.GetRegisterPage)
		r.Post("/register", reg.PostRegister)
		r.Post("/logout", login.PostLogout)

		r.Group(func(r chi.Router) {
			r.Use(m.SessionAuthMiddleware(sessionStore))

			r.Get("/", handlers.NewHomeHandler(sessionStore).ServeHTTP)

			r.Route("/produtos", func(r chi.Router) {
				r.Get("/", product.GetProductPage)
				r.Get("/novo", product.GetProductForm)
				r.Get("/{id}", product.GetEditPage)
				r.Post("/", product.PostNewProduct)
				r.Post("/{id}", product.UpdateProduct)
				r.Delete("/{id}", product.DeleteProduct)

				r.Get("/pricetable", product.GetTablePage)
				r.Post("/pricetable", product.CreatePriceTable)
				r.Delete("/pricetable/{id}", product.DeletePriceTable)
			})

			r.Route("/contacts", func(r chi.Router) {
				r.Get("/", contact.GetContactsPage)
				r.Get("/novo", contact.GetContactsForm)
				r.Get("/{id}", contact.GetEditPage)
				r.Post("/", contact.PostNewContact)
				r.Post("/{id}", contact.Update)
				r.Post("/{id}/create-link", contact.CreateLink)
			})

			r.Route("/pedidos", func(r chi.Router) {
				r.Get("/", order.GetOrdersPage)
				r.Get("/novo", order.GetNewOrderPage)
				r.Get("/produtos", order.SearchProductsForOrder)
				r.Post("/", order.PostNewOrder)
				r.Get("/{id}", order.GetOrderPage)
			})
		})
	})
}
