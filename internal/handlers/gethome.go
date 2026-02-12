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

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const website_name = "EZ"
	var is_hxRequest = r.Header.Get("HX-Request") == "true"

	user := middleware.GetUser(r.Context())
	loggedIn := user != nil

	email := ""
	var id uint
	if user != nil {
		email = user.Email
		id = user.ID
	}

	if !loggedIn {
		if is_hxRequest {
			err := templates.GuestIndex().Render(r.Context(), w)
			if err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
				return
			}
			return
		}

		err := templates.Layout(templates.GuestIndex(), website_name, false, "").Render(r.Context(), w)

		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
		return
	}

	if is_hxRequest {
		err := templates.Index(email, id).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
		return
	}

	err := templates.Layout(templates.Index(email, id), website_name, true, email).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
