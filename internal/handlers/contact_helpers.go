package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/DTineli/ez/internal/forms"
	"github.com/DTineli/ez/internal/store"
)

func mapContactToForm(c *store.Contact) *forms.Form {
	form := forms.New(nil)

	form.Set("id", strconv.Itoa(int(c.ID)))
	form.Set("name", c.Name)
	form.Set("trade_name", c.TradeName)
	form.Set("contact_type", string(c.ContactType))

	form.Set("document_type", c.DocumentType)
	form.Set("document", c.Document)
	form.Set("ie", c.IE)

	form.Set("email", c.Email)
	form.Set("phone", c.Phone)

	form.Set("zipcode", c.ZipCode)
	form.Set("street", c.Street)
	form.Set("number", c.Number)
	form.Set("complement", c.Complement)
	form.Set("neighborhood", c.Neighborhood)
	form.Set("city", c.City)
	form.Set("uf", c.UF)

	form.Set("price_table", joinPriceTableIDs(c.PriceTables))

	form.Set("invite_link", c.InviteLink)

	return form
}

func validateContactForm(r *http.Request) (*forms.Form, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	form := forms.New(r.PostForm)

	form.Set(
		"document",
		strings.NewReplacer(".", "", "/", "", "-", "").
			Replace(form.Get("document")),
	)
	form.Set(
		"phone",
		strings.NewReplacer("(", "", ")", "", " ", "", "-", "").
			Replace(form.Get("phone")),
	)

	form.Required(
		"name",
		"trade_name",
		"phone",
		"contact_type",
		"document_type",
	)

	form.MaxLength("name", 255)
	form.MaxLength("trade_name", 255)

	form.IsInt("document")
	form.IsInt("ie")
	form.IsInt("zipcode")

	form.IsEmail("email")

	return form, nil
}

func GetPagination(r *http.Request) store.Pagination {
	page := 1
	perPage := 10
	if strPage := r.URL.Query().Get("page"); strPage != "" {
		if p, err := strconv.Atoi(strPage); err == nil && p > 0 {
			page = p
		}
	}

	return store.Pagination{Page: page, PerPage: perPage}
}

func joinPriceTableIDs(tables []store.PriceTable) string {
	ids := make([]string, 0, len(tables))
	for _, t := range tables {
		ids = append(ids, strconv.Itoa(int(t.ID)))
	}
	return strings.Join(ids, ",")
}

func convertPriceTable(tables []string) []store.PriceTable {
	var priceTables []store.PriceTable

	for _, idStr := range tables {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			continue
		}
		priceTables = append(priceTables, store.PriceTable{ID: uint(id)})
	}

	return priceTables
}
