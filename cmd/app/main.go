package main

import (
	"net/http"

	"github.com/DTineli/ez/cmd/internal/web/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func HelloWorld(w http.ResponseWriter, r *http.Request) {
	main_page := views.MainPage()

	main_page.Render(r.Context(), w)
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", HelloWorld)

	// r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte("welcome"))
	// })

	http.ListenAndServe(":3000", r)
}
