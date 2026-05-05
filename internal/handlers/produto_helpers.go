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

	form.IsEAN("ean")

	form.IsFloat("weight")
	form.IsFloat("height")
	form.IsFloat("width")
	form.IsFloat("length")

	return form, nil
}

func mapProductToForm(p *store.Product) *forms.Form {
	form := forms.New(nil)

	form.Set("ID", strconv.Itoa(int(p.ID)))
	form.Set("name", p.Name)
	form.Set("sku", p.SKU)
	form.Set("uom", string(p.UOM))
	form.Set("description", p.FullDescription)
	form.Set("weight", strconv.FormatFloat(p.Weight, 'f', 3, 64))
	form.Set("height", strconv.FormatFloat(p.Height, 'f', 3, 64))
	form.Set("width", strconv.FormatFloat(p.Width, 'f', 3, 64))
	form.Set("length", strconv.FormatFloat(p.Length, 'f', 3, 64))

	ncm := p.NCM
	if len(ncm) == 8 {
		ncm = ncm[:4] + "." + ncm[4:6] + "." + ncm[6:]
	}
	form.Set("ncm", ncm)

	return form
}
