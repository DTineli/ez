package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/DTineli/ez/internal/forms"
	"github.com/DTineli/ez/internal/store"
)

func validateProductForm(r *http.Request) (*forms.Form, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	form := forms.New(r.PostForm)

	form.Required("name", "sku")
	form.MaxLength("name", 255)
	form.MinLength("name", 4)

	form.MaxLength("uom", 2)
	form.MinLength("uom", 2)

	form.MaxLength("description", 15000)

	form.MaxLength("sku", 25)
	form.MinLength("sku", 4)

	form.IsInt("ean")
	form.Set("ncm", strings.ReplaceAll(form.Get("ncm"), ".", ""))
	form.IsInt("ncm")

	return form, nil
}

func mapProductToForm(p *store.Product) *forms.Form {
	form := forms.New(nil)

	form.Set("ID", strconv.Itoa(int(p.ID)))
	form.Set("name", p.Name)
	form.Set("sku", p.SKU)
	form.Set("uom", string(p.UOM))
	form.Set("description", p.FullDescription)
	form.Set("ean", p.EAN)

	ncm := p.NCM
	if len(ncm) == 8 {
		ncm = ncm[:4] + "." + ncm[4:6] + "." + ncm[6:]
	}
	form.Set("ncm", ncm)

	return form
}
