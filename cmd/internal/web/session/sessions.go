package session

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var Store = sessions.NewCookieStore([]byte("CHAVE-BEM-SECRETA-AQUI"))

func Configure() {
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60 * 24, // 1 dia
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		// Secure: true,
	}
}
