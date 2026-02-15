package handlers

import (
	b64 "encoding/base64"
	"net/http"
	"strconv"
	"strings"

	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
	"golang.org/x/crypto/bcrypt"
)

type LoginHandler struct {
	userStore    store.UserStore
	sessionStore store.SessionStore
	cookieName   string
}

func NewLoginHandler(userStore store.UserStore, sessionStore store.SessionStore, cookieName string) *LoginHandler {
	return &LoginHandler{
		userStore:    userStore,
		sessionStore: sessionStore,
		cookieName:   cookieName,
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

	user, err := h.userStore.GetUser(email)
	if err != nil || user == nil {
		writeLoginError(r, w, "Email ou senha incorretos.")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		writeLoginError(r, w, "Email ou senha incorretos.")
		return
	}

	sess, err := h.sessionStore.CreateSession(&store.Session{UserID: user.ID})
	if err != nil {
		writeLoginError(r, w, "Erro ao criar sessão. Tente novamente.")
		return
	}

	cookieValue := b64.StdEncoding.EncodeToString(
		[]byte(sess.SessionID + ":" + strconv.FormatUint(uint64(sess.UserID), 10)),
	)

	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieName,
		Value:    cookieValue,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 3600, // 7 dias
	})

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
