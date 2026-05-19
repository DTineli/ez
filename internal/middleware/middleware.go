package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/DTineli/ez/internal/store"
)

type contextKey string
type SessionInfoContextKey string

const slugKey contextKey = "slug"

var SessionInfoKey SessionInfoContextKey = "info"

func TextHTMLMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

func ResolveSlug(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		slug := strings.Split(host, ".")[0]
		if slug == r.Host {
			slug = ""
		}
		ctx := context.WithValue(r.Context(), slugKey, slug)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func SessionAuthMiddleware(
	sessionStore store.SessionStore,
) func(http.Handler) http.Handler {
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

func GetSlugFromContext(r *http.Request) string {
	slug, _ := r.Context().Value(slugKey).(string)
	return slug
}

func GetSessionFromContext(r *http.Request) *store.Session {
	sess, _ := r.Context().Value(SessionInfoKey).(*store.Session)
	return sess
}

func parseUrl(url string) store.AccessType {
	if strings.HasPrefix(url, "/admin") {
		return store.AccessAdmin
	}
	return store.AccessCustomer
}

func redirectTo(
	w http.ResponseWriter,
	r *http.Request,
	ambiente store.AccessType,
) {
	if ambiente == store.AccessCustomer {
		http.Redirect(w, r, "/client/login", http.StatusFound)
	} else {
		http.Redirect(w, r, "/admin/login", http.StatusFound)
	}
}
