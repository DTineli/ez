package handlers

import (
	"fmt"
	"net/http"
	"strconv"
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

func formUint(r *http.Request, key string) (uint64, error) {
	v, err := strconv.ParseUint(r.FormValue(key), 10, 64)
	if err != nil || v == 0 {
		return 0, fmt.Errorf("%s inválido", key)
	}
	return v, nil
}

func formPosInt(r *http.Request, key string) (int, error) {
	v, err := strconv.Atoi(r.FormValue(key))
	if err != nil || v <= 0 {
		return 0, fmt.Errorf("%s inválido", key)
	}
	return v, nil
}

// maxVariantCombos limita o produto cartesiano de atributos pra evitar
// explosão combinatória acidental num único submit.
const maxVariantCombos = 200

type attrAxis struct {
	Name   string
	Values []string
}

// parseVariantAxes lê os eixos de atributo do form: attr_name_N (nome, um por
// eixo) e attr_values_N (repetido, um campo por valor escolhido no eixo N).
// Eixos sem nome ou sem nenhum valor são ignorados. Retorna erro se o mesmo
// nome de atributo aparecer em mais de um eixo.
func parseVariantAxes(r *http.Request) ([]attrAxis, error) {
	seen := map[string]bool{}
	var axes []attrAxis
	for key, vals := range r.Form {
		if !strings.HasPrefix(key, "attr_name_") || len(vals) == 0 {
			continue
		}
		name := strings.TrimSpace(vals[0])
		if name == "" {
			continue
		}
		idxStr := strings.TrimPrefix(key, "attr_name_")

		var values []string
		seenValues := map[string]bool{}
		for _, raw := range r.Form["attr_values_"+idxStr] {
			value := strings.TrimSpace(raw)
			if value == "" {
				continue
			}
			valueLower := strings.ToLower(value)
			if seenValues[valueLower] {
				continue
			}
			seenValues[valueLower] = true
			values = append(values, value)
		}
		if len(values) == 0 {
			continue
		}

		nameLower := strings.ToLower(name)
		if seen[nameLower] {
			return nil, fmt.Errorf("atributo duplicado: %s", name)
		}
		seen[nameLower] = true
		axes = append(axes, attrAxis{Name: name, Values: values})
	}
	return axes, nil
}

// cartesianCombos gera o produto cartesiano dos valores de cada eixo. Cada
// combo resultante é uma tupla ordenada com um valor por eixo, na mesma
// ordem de axes.
func cartesianCombos(axes []attrAxis) [][]string {
	if len(axes) == 0 {
		return nil
	}
	combos := [][]string{{}}
	for _, axis := range axes {
		next := make([][]string, 0, len(combos)*len(axis.Values))
		for _, combo := range combos {
			for _, value := range axis.Values {
				extended := make([]string, len(combo), len(combo)+1)
				copy(extended, combo)
				next = append(next, append(extended, value))
			}
		}
		combos = next
	}
	return combos
}
