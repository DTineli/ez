package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/DTineli/ez/internal/templates"
	"github.com/a-h/templ"
)

// slugFromHost extrai o primeiro segmento do hostname, ignorando a porta.
// "company.localhost:4000" → "company"
// "localhost:4000"         → "localhost"
func slugFromHost(host string) string {
	hostname := strings.Split(host, ":")[0]
	return strings.Split(hostname, ".")[0]
}

func Render(c templ.Component, r *http.Request, w http.ResponseWriter) error {
	if r.Header.Get("HX-Request") == "true" {
		return c.Render(r.Context(), w)
	}

	return templates.
		Layout(c, "Ez", true, "").
		Render(r.Context(), w)
}

func ShowToast(w http.ResponseWriter, message string, toastType string) {
	w.Header().Set("HX-Trigger", fmt.Sprintf(`{
	"showToast": {
		"type": "%v",
		"message": "%v"
	}
}`, toastType, message))
	w.WriteHeader(http.StatusOK)
}

func isDuplicateError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") ||
		strings.Contains(msg, "Duplicate")
}
