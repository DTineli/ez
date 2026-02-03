package main

import (
	"net/http"

	"github.com/DTineli/ez/cmd/internal/web/session"
	"github.com/DTineli/ez/cmd/internal/web/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func HelloWorld(w http.ResponseWriter, r *http.Request) {
	main_page := views.MainPage()

	main_page.Render(r.Context(), w)
}

func login(w http.ResponseWriter, r *http.Request) {
	session, err := session.Store.Get(r, "session")
	if err != nil {
		http.Error(w, "Erro na sessao", http.StatusInternalServerError)
	}

	session.Values["user_name"] = "Nelso"
	session.Values["email"] = r.FormValue("email")

	session.Save(r, w)
	w.Header().Set("HX-Redirect", "/dashboard")
}

func get_dashboard(w http.ResponseWriter, r *http.Request) {
	sess, err := session.Store.Get(r, "session")
	if err != nil {
		http.Error(w, "Erro na sessao", http.StatusInternalServerError)
	}

	email, _ := sess.Values["email"].(string)
	if email == "" {
		w.Header().Set("HX-Redirect", "/")
		return
	}

	dashboard_page := views.Nome_email("Nelso", email)

	dashboard_page.Render(r.Context(), w)
}

func main() {
	session.Configure()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", HelloWorld)
	r.Post("/login", login)
	r.Get("/dashboard", get_dashboard)

	// r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte("welcome"))
	// })

	http.ListenAndServe(":3000", r)
}
