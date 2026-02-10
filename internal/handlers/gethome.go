package handlers

import (
	"net/http"

	"github.com/DTineli/ez/internal/middleware"
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

	user := middleware.GetUser(r.Context())
	loggedIn := user != nil
	email := ""
	if user != nil {
		email = user.Email
	}

	if !loggedIn {
		err := templates.Layout(templates.GuestIndex(), website_name, false, "").Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
		return
	}

	err := templates.Layout(templates.Index(email), website_name, true, email).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
