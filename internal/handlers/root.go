package handlers

import (
	"fmt"
	"net/http"

	"github.com/DTineli/ez/internal/middleware"
)

type RootHandler struct{}

func NewRootHandler() *RootHandler {
	return &RootHandler{}
}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slug := middleware.GetSlugFromContext(r)
	sess := middleware.GetSessionFromContext(r)

	if sess != nil {
		fmt.Println("Sess - ", sess.UserAccessType)
	}

	if slug == "" {
		w.Write([]byte("<h1>Sem Slug</h1>"))
		return
	}

	http.Redirect(w, r, "/client/login", http.StatusFound)
}
