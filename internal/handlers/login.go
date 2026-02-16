package handlers

import (
	"net/http"
	"strings"

	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
	"golang.org/x/crypto/bcrypt"
)

type LoginHandler struct {
	userStore    store.UserStore
	sessionStore store.SessionStore
	tenantStore  store.TenantStore
	cookieName   string
}

type LoginHandlerParams struct {
	UserStore    store.UserStore
	SessionStore store.SessionStore
	TenantStore  store.TenantStore
	CookieName   string
}

func NewLoginHandler(params LoginHandlerParams) *LoginHandler {
	return &LoginHandler{
		userStore:    params.UserStore,
		sessionStore: params.SessionStore,
		tenantStore:  params.TenantStore,
		cookieName:   params.CookieName,
	}
}

func (h *LoginHandler) GetLoginPage(w http.ResponseWriter, r *http.Request) {
	var is_hxRequest = r.Header.Get("HX-Request") == "true"

	if is_hxRequest {
		err := templates.LoginPage().Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
		return
	}

	err := templates.Layout(templates.LoginPage(), "Login", false, "").Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func (h *LoginHandler) PostLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeLoginError(r, w, "Dados inválidos.")
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	if email == "" {
		writeLoginError(r, w, "Email é obrigatório.")
		return
	}
	if password == "" {
		writeLoginError(r, w, "Senha é obrigatória.")
		return
	}

	//TODO: getUserWithTenant
	user, err := h.userStore.GetUser(email)
	if err != nil || user == nil {
		writeLoginError(r, w, "Email ou senha incorretos.")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		writeLoginError(r, w, "Email ou senha incorretos.")
		return
	}

	tenant, err := h.tenantStore.GetTenantByID(user.TenantID)
	if err != nil {
		writeLoginError(r, w, "Erro ao criar sessão. Tente novamente.")
		return
	}

	// TODO: Se ele ta no slug errado troca ou da erro ?
	if tenant.Slug != strings.Split(r.Host, ".")[0] {
		// w.Header().Set(HXRedirect, fmt.Sprintf("http://%s.localhost:4000", tenant.Slug))
		writeLoginError(r, w, "slug diferente")
		return
	}

	err = h.sessionStore.CreateSession(r, w, store.Session{
		UserID:     user.ID,
		UserEmail:  user.Email,
		TenantID:   tenant.ID,
		TenantSlug: tenant.Slug,
	})

	if err != nil {
		writeLoginError(r, w, "Erro ao criar sessão. Tente novamente.")
		return
	}

	w.Header().Set(HXRedirect, "/")
	w.WriteHeader(http.StatusOK)
}

func (h *LoginHandler) PostLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	w.Header().Set(HXRedirect, "/")
	w.WriteHeader(http.StatusOK)
}

func writeLoginError(r *http.Request, w http.ResponseWriter, message string) {
	// w.WriteHeader(http.StatusUnprocessableEntity)
	_ = templates.LoginErrors(message).Render(r.Context(), w)
}
