package handlers

import (
	"net/http"

	"github.com/DTineli/ez/internal/templates"
)

func (p *ProductHandler) GetTablePage(w http.ResponseWriter, r *http.Request) {
	Render(templates.PriceTablePage(), r, w)
}

func (p *ProductHandler) CreatePriceTable(w http.ResponseWriter, r *http.Request) {

}
