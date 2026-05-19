package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

var ErrInvalidForm = errors.New("invalid form")

type priceTableDTO struct {
	Name       string
	Percentage float64
}

func formToPriceTable(r *http.Request) (*priceTableDTO, map[string]string) {
	errs := make(map[string]string)

	if err := r.ParseForm(); err != nil {
		errs["tec"] = err.Error()
		return nil, errs
	}

	name := r.FormValue("name")
	if name == "" {
		errs["name"] = fmt.Errorf("%w: nome vazio", ErrInvalidForm).Error()
	}

	percentageStr := r.FormValue("percentage")
	percentage, err := strconv.ParseFloat(percentageStr, 64)
	if err != nil {
		errs["percentage"] = fmt.Errorf("%w: percentage inválido", ErrInvalidForm).
			Error()
	}

	if percentage < 0 {
		errs["percentage"] = fmt.Errorf("%w: percentage negativo", ErrInvalidForm).
			Error()
	}

	return &priceTableDTO{
		Name:       name,
		Percentage: percentage,
	}, errs
}
