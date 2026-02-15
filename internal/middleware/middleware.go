package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/DTineli/ez/internal/store"
)

type UserContextKey string
type TenantContextKey string
type SessionInfoContextKey string

var UserKey UserContextKey = "user"
var TenantKey TenantContextKey = "tenant"
var SessionInfoKey SessionInfoContextKey = "info"

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

func GetSessionInfo(ctx context.Context) *store.Session {
	sessionInfo := ctx.Value(SessionInfoKey)
	u, ok := sessionInfo.(*store.Session)
	if !ok {
		return nil
	}
	return u
}

func CheckTenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenant := strings.Split(r.Host, ".")[0]

		ctx := context.WithValue(r.Context(), TenantKey, tenant)
		next.ServeHTTP(w, r.WithContext(ctx))

		next.ServeHTTP(w, r)
	})

}

func (m *AuthMiddleware) AddSessionInfoToContext(next http.Handler) http.Handler {
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

		sessionID := sessionCookie.Value
		sessionInfo, err := m.sessionStore.GetSessionInfo(sessionID)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		if sessionInfo == nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), SessionInfoKey, sessionInfo)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
