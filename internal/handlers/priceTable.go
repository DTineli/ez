package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	m "github.com/DTineli/ez/internal/middleware"
	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
)

var ErrInvalidForm = errors.New("invalid form")

type priceTableDTO struct {
	Name       string
	Percentage float64
}

func formToPriceTable(r *http.Request) (*priceTableDTO, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err // erro técnico
	}

	name := r.FormValue("name")
	if name == "" {
		return nil, fmt.Errorf("%w: Nome vazio", ErrInvalidForm)
	}

	percentageStr := r.FormValue("percentage")
	percentage, err := strconv.ParseFloat(percentageStr, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: percentage inválido", ErrInvalidForm)
	}

	if percentage < 0 {
		return nil, fmt.Errorf("%w: percentage negativo", ErrInvalidForm)
	}

	return &priceTableDTO{
		Name:       name,
		Percentage: percentage,
	}, nil
}

func (p *ProductHandler) GetTablePage(w http.ResponseWriter, r *http.Request) {
	Render(templates.PriceTablePage(), r, w)
}

func (p *ProductHandler) CreatePriceTable(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)

	dto, err := formToPriceTable(r)

	if err != nil {
		if errors.Is(err, ErrInvalidForm) {
			// erro de validação → volta para o form
			Render(templates.PriceTablePage(), r, w)
			return
		}
		// erro técnico → 500
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err = p.priceTableStore.CreatePriceTable(&store.PriceTable{
		Name:       dto.Name,
		Percentage: dto.Percentage,

		TenantID: sess.TenantID,
	}); err != nil {
		Render(templates.PriceTablePage(), r, w)
		return
	}

	Render(templates.PriceTablePage(), r, w)
}
