package handlers

import (
	"net/http"
	"strings"

	"github.com/DTineli/ez/internal/middleware"
	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
)

const HXRedirect = "HX-Redirect"

type RegisterHandler struct {
	userStore store.UserStore
}

func NewRegisterHandler(userStore store.UserStore) *RegisterHandler {
	return &RegisterHandler{userStore: userStore}
}

func NewRegisterHandlerWithService() *RegisterHandler {
	return &RegisterHandler{userStore: nil}
}

func (h *RegisterHandler) GetRegisterPage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	loggedIn := user != nil
	email := ""
	if user != nil {
		email = user.Email
	}
	err := templates.Layout(templates.RegisterPage(), "Criar conta", loggedIn, email).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func (h *RegisterHandler) PostRegister(w http.ResponseWriter, r *http.Request) {
	if h.userStore == nil {
		http.Error(w, "user store not configured", http.StatusInternalServerError)
		return
	}
	if err := r.ParseForm(); err != nil {
		writeRegisterError(r, w, "Dados inválidos.")
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	if name == "" {
		writeRegisterError(r, w, "Nome é obrigatório.")
		return
	}
	if email == "" {
		writeRegisterError(r, w, "Email é obrigatório.")
		return
	}
	if len(password) < 6 {
		writeRegisterError(r, w, "Senha deve ter no mínimo 6 caracteres.")
		return
	}

	err := h.userStore.CreateUser(store.UserDTO{
		Name:     name,
		Email:    email,
		Password: password,
	})
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "Duplicate") {
			writeRegisterError(r, w, "Este email já está em uso.")
			return
		}
		writeRegisterError(r, w, "Erro ao criar conta. Tente novamente.")
		return
	}

	w.Header().Set(HXRedirect, "/")
	w.WriteHeader(http.StatusOK)
}

func writeRegisterError(r *http.Request, w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusUnprocessableEntity)
	_ = templates.RegisterErrors(message).Render(r.Context(), w)
}
