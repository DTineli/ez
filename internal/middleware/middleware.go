package middleware

import (
	"context"
	"net/http"

	"github.com/DTineli/ez/internal/store"
)

func GetSessionFromContext(r *http.Request) *store.Session {
	return r.Context().Value(SessionInfoKey).(*store.Session)
}

func SessionAuthMiddleware(store store.SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, err := store.GetSessionInfo(r)

			if err != nil {
				http.Redirect(w, r, "/login", http.StatusFound)
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
