package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
)

const HXRedirect = "HX-Redirect"
const MIN_LEN_PASSWD = 4

type RegisterHandler struct {
	userStore   store.UserStore
	tenantStore store.TenantStore
}

func NewRegisterHandler(userStore store.UserStore, tenantStore store.TenantStore) *RegisterHandler {
	return &RegisterHandler{userStore: userStore, tenantStore: tenantStore}
}

func NewRegisterHandlerWithService() *RegisterHandler {
	return &RegisterHandler{userStore: nil}
}

func (h *RegisterHandler) GetRegisterPage(w http.ResponseWriter, r *http.Request) {
	var is_hxRequest = r.Header.Get("HX-Request") == "true"
	if is_hxRequest {
		err := templates.RegisterPage().Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
		return
	}

	err := templates.Layout(templates.RegisterPage(), "Criar conta", false, "").Render(r.Context(), w)
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
	slug := strings.TrimSpace(r.FormValue("slug"))
	password := r.FormValue("password")
	password_confirmation := r.FormValue("password_confirmation")

	if name == "" {
		writeRegisterError(r, w, "Nome é obrigatório.")
		return
	}

	if email == "" {
		writeRegisterError(r, w, "Email é obrigatório.")
		return
	}

	if slug == "" {
		writeRegisterError(r, w, "Slug é obrigatório.")
		return
	}

	if password != password_confirmation {
		writeRegisterError(r, w, "senhas nao batem")
		return
	}

	if len(password) < MIN_LEN_PASSWD {
		writeRegisterError(r, w, fmt.Sprintf("Senha deve ter no mínimo %v caracteres.", MIN_LEN_PASSWD))
		return
	}

	tenantID, err := h.tenantStore.CreateTenant(store.Tenant{
		Slug:     slug,
		Document: "",
	})

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "Duplicate") {
			writeRegisterError(r, w, "Este Slug já está em uso.")
			return
		}
		writeRegisterError(r, w, "Erro ao criar conta. Tente novamente.")
		return
	}

	err = h.userStore.CreateUser(store.User{
		Name:     name,
		Email:    email,
		TenantID: tenantID,
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

	url := fmt.Sprintf("http://%s.%s", slug, r.Host)

	w.Header().Set(HXRedirect, url)
	w.WriteHeader(http.StatusOK)
}

func writeRegisterError(r *http.Request, w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// w.WriteHeader(http.StatusUnprocessableEntity)

	_ = templates.RegisterErrors(message).Render(r.Context(), w)
}
