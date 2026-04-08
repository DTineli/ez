package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/DTineli/ez/internal/store"
)

func GetSessionFromContext(r *http.Request) *store.Session {
	return r.Context().Value(SessionInfoKey).(*store.Session)
}

func parseUrl(url string) store.AccessType {
	if strings.HasPrefix(url, "/admin") {
		return store.AccessAdmin
	}

	return store.AccessCustomer
}

func redirectTo(w http.ResponseWriter, r *http.Request, ambiente store.AccessType) {
	if ambiente == store.AccessCustomer {
		http.Redirect(w, r, "/client/login", http.StatusFound)
	} else {
		http.Redirect(w, r, "/admin/login", http.StatusFound)
	}
}

func SessionAuthMiddleware(sessionStore store.SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, err := sessionStore.GetSessionInfo(r)

			ambiente := parseUrl(r.URL.Path)

			if err != nil {
				redirectTo(w, r, ambiente)
				return
			}

			if ambiente != sess.UserAccessType {
				redirectTo(w, r, ambiente)
				return
			}

			ctx := context.WithValue(r.Context(), SessionInfoKey, sess)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type SessionInfoContextKey string

var SessionInfoKey SessionInfoContextKey = "info"

func TextHTMLMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}
