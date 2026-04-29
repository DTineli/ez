package handlers

import (
	"fmt"
	"net/http"
	"strings"

	m "github.com/DTineli/ez/internal/middleware"
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

	sess := m.GetSessionFromContext(r)
	email := ""
	if sess != nil {
		email = sess.UserEmail
	}

	return templates.
		Layout(c, "Ez", true, email).
		Render(r.Context(), w)
}

// HXLocation faz client-side navigate via HTMX sem reload completo.
// Swapa apenas #main-content, preservando sidebar e header (Alpine.js state).
// Aceita msg/type opcionais pra toast junto com o navigate.
func HXLocation(w http.ResponseWriter, path string, toast ...string) {
	w.Header().Set("HX-Location", fmt.Sprintf(`{"path":%q,"target":"#main-content","swap":"innerHTML"}`, path))
	if len(toast) == 2 {
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast":{"type":%q,"message":%q}}`, toast[1], toast[0]))
	}
	w.WriteHeader(http.StatusOK)
}

func ShowToast(w http.ResponseWriter, message string, toastType string) {
	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast":{"type":%q,"message":%q}}`, toastType, message))
	w.WriteHeader(http.StatusOK)
}

func isDuplicateError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") || // SQLite
		strings.Contains(msg, "Duplicate") || // MySQL
		strings.Contains(msg, "duplicate key value violates unique constraint") // PostgreSQL
}

type variantAttrInput struct {
	Name  string
	Value string
}

// parseVariantAttributes lê pares attr_name_X/attr_value_X do form.
// Retorna erro se o mesmo nome aparecer mais de uma vez.
func parseVariantAttributes(r *http.Request) ([]variantAttrInput, error) {
	seen := map[string]bool{}
	var attrs []variantAttrInput
	for key, vals := range r.Form {
		if !strings.HasPrefix(key, "attr_name_") || len(vals) == 0 {
			continue
		}
		name := strings.TrimSpace(vals[0])
		if name == "" {
			continue
		}
		idxStr := strings.TrimPrefix(key, "attr_name_")
		value := strings.TrimSpace(r.FormValue("attr_value_" + idxStr))
		if value == "" {
			continue
		}
		nameLower := strings.ToLower(name)
		if seen[nameLower] {
			return nil, fmt.Errorf("atributo duplicado: %s", name)
		}
		seen[nameLower] = true
		attrs = append(attrs, variantAttrInput{Name: name, Value: value})
	}
	return attrs, nil
}
