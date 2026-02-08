package handlers

import (
	"net/http"

	"github.com/DTineli/ez/internal/middleware"
	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
)

type HomeHandler struct{}

func NewHomeHandler() *HomeHandler {
	return &HomeHandler{}
}

func (h *HomeHandler) GetMainPage(w http.ResponseWriter, r *http.Response) {

}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const website_name = "EZ"

	user, ok := r.Context().Value(middleware.UserKey).(*store.User)

	if !ok {
		c := templates.GuestIndex()

		err := templates.Layout(c, website_name).Render(r.Context(), w)

		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}

		return
	}

	c := templates.Index(user.Email)
	err := templates.Layout(c, website_name).Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
