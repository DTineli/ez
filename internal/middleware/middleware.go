package middleware

import (
	"context"
	b64 "encoding/base64"
	"net/http"
	"strings"

	"github.com/DTineli/ez/internal/store"
)

var UserKey UserContextKey = "user"

type UserContextKey string
type AuthMiddleware struct {
	sessionStore      store.SessionStore
	sessionCookieName string
}

func TextHTMLMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

func NewAuthMiddleware(sessionStore store.SessionStore, sessionCookieName string) *AuthMiddleware {
	return &AuthMiddleware{
		sessionStore:      sessionStore,
		sessionCookieName: sessionCookieName,
	}
}

func GetUser(ctx context.Context) *store.User {
	user := ctx.Value(UserKey)
	if user == nil {
		return nil
	}
	u, ok := user.(*store.User)
	if !ok {
		return nil
	}
	return u
}

func (m *AuthMiddleware) AddUserToContext(next http.Handler) http.Handler {
	if m == nil || m.sessionStore == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, err := r.Cookie(m.sessionCookieName)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		if sessionCookie == nil || sessionCookie.Value == "" {
			next.ServeHTTP(w, r)
			return
		}

		decodedValue, err := b64.StdEncoding.DecodeString(sessionCookie.Value)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		splitValue := strings.Split(string(decodedValue), ":")
		if len(splitValue) != 2 {
			next.ServeHTTP(w, r)
			return
		}

		sessionID := strings.TrimSpace(splitValue[0])
		userID := strings.TrimSpace(splitValue[1])
		if sessionID == "" || userID == "" {
			next.ServeHTTP(w, r)
			return
		}

		user, err := m.sessionStore.GetUserFromSession(sessionID, userID)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		if user == nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), UserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
