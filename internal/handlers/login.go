package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
	"golang.org/x/crypto/bcrypt"
)

type LoginHandler struct {
	userStore    store.UserStore
	contactStore store.ContactStore
	tenantStore  store.TenantStore

	sessionStore store.SessionStore
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

	// err := Render(templates.LoginPage(), r, w)
	err := templates.LoginPage().Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func (h *LoginHandler) PostLoginHandler(accessType store.AccessType) http.HandlerFunc {
	switch accessType {
	case store.AccessAdmin:
		return h.adminLogin

	case store.AccessCustomer:
		return h.customerLogin

	default:
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "invalid access type", http.StatusInternalServerError)
		}
	}
}

func (h *LoginHandler) customerLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeLoginError(r, w, "Dados inválidos.")
		return
	}

	phone_number := strings.TrimSpace(r.FormValue("phone_number"))
	password := r.FormValue("password")

	if phone_number == "" {
		writeLoginError(r, w, "Fone é obrigatório.")
		return
	}
	if password == "" {
		writeLoginError(r, w, "Senha é obrigatória.")
		return
	}

	//TODO: getUserWithTenant
	user, err := h.userStore.GetUser("")
	if err != nil || user == nil {
		writeLoginError(r, w, "Falha no Login")
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
		writeLoginError(r, w, "slug diferente")
		return
	}

	err = h.sessionStore.CreateSession(r, w, store.Session{
		UserAccessType: store.AccessAdmin,
		UserID:         user.ID,
		UserEmail:      user.Email,
		TenantID:       tenant.ID,
		TenantSlug:     tenant.Slug,
	})

	if err != nil {
		writeLoginError(r, w, "Erro ao criar sessão. Tente novamente.")
		return
	}

	w.Header().Set(HXRedirect, "/admin/")
	w.WriteHeader(http.StatusOK)

}

func (h *LoginHandler) adminLogin(w http.ResponseWriter, r *http.Request) {
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
		writeLoginError(r, w, "slug diferente")
		return
	}

	err = h.sessionStore.CreateSession(r, w, store.Session{
		UserAccessType: store.AccessAdmin,
		UserID:         user.ID,
		UserEmail:      user.Email,
		TenantID:       tenant.ID,
		TenantSlug:     tenant.Slug,
	})

	if err != nil {
		writeLoginError(r, w, "Erro ao criar sessão. Tente novamente.")
		return
	}

	w.Header().Set(HXRedirect, "/admin/")
	w.WriteHeader(http.StatusOK)
}

func (h *LoginHandler) PostLogout(w http.ResponseWriter, r *http.Request) {

	err := h.sessionStore.DeleteSession(r, w)
	if err != nil {
		fmt.Println(err)
		return
	}

	w.Header().Set(HXRedirect, "/login")
	w.WriteHeader(http.StatusOK)
}

func writeLoginError(r *http.Request, w http.ResponseWriter, message string) {
	// w.WriteHeader(http.StatusUnprocessableEntity)
	_ = templates.LoginErrors(message).Render(r.Context(), w)
}
