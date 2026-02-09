package handlers

import (
	"net/http"

	"github.com/DTineli/ez/internal/templates"
)

type RegisterService struct {
}

type RegisterHandler struct {
	service *RegisterService
}

func NewRegisterHandlerWithService() *RegisterHandler {
	return &RegisterHandler{
		service: &RegisterService{},
	}
}

func (l *RegisterHandler) GetRegisterPage(w http.ResponseWriter, r *http.Request) {
	err := templates.Layout(templates.RegisterPage(), "My Page").Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
