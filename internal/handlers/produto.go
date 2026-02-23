package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/DTineli/ez/internal/forms"
	m "github.com/DTineli/ez/internal/middleware"
	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
)

type ProductHandler struct {
	productStore store.ProductStore
}

func NewProductHandler(db store.ProductStore) *ProductHandler {
	return &ProductHandler{
		productStore: db,
	}
}

func Render(templ templ.Component, r *http.Request, w http.ResponseWriter) error {
	var is_hxRequest = r.Header.Get("HX-Request") == "true"
	if is_hxRequest {
		return templ.Render(r.Context(), w)
	}

	return templates.Layout(
		templ,
		"Ez",
		true, //TODO: verificar se ta logado
		"",
	).Render(r.Context(), w)
}

func (p *ProductHandler) GetProductForm(w http.ResponseWriter, r *http.Request) {
	err := Render(templates.ProductForm(forms.New(r.PostForm), false), r, w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func (p *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {

}

func (p *ProductHandler) PostNewProduct(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)

	form.Required("name", "sku")

	form.MaxLength("name", 255)
	form.MinLength("name", 4)

	form.MaxLength("uom", 2)
	form.MinLength("uom", 2)

	form.MaxLength("description", 255)

	form.MaxLength("sku", 25)
	form.MinLength("sku", 4)

	costPrice := form.IsFloat("cost_price")
	weight := form.IsFloat("weight")
	height := form.IsFloat("height")
	width := form.IsFloat("width")
	length := form.IsFloat("Length")

	currentStock := form.IsInt("current_stock")
	minimumStock := form.IsInt("minimum_stock")
	form.IsInt("ean")

	if !form.Valid() {
		err := Render(templates.ProductForm(form, false), r, w)
		if err != nil {
			http.Error(w, "Error Creating Product", http.StatusInternalServerError)
		}
		return
	}

	sess := m.GetSessionFromContext(r)

	//TODO: Verificar sku duplicado

	err := p.productStore.CreateProduct(&store.Product{
		TenantID:        sess.TenantID,
		SKU:             form.Get("sku"),
		Name:            form.Get("name"),
		FullDescription: form.Get("description"),
		Status:          true, //TODO: Persistir isso daqui
		UOM:             store.UOM(form.Get("uom")),
		EAN:             form.Get("ean"),
		NCM:             form.Get("ncm"),
		CostPrice:       costPrice,
		WidthCm:         width,
		Weight:          weight,
		HeightCm:        height,
		LengthCm:        length,
		MinimumStock:    minimumStock,
		CurrentStock:    currentStock,
	})

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "Duplicate") {
			form.Errors.Add("sku", "Este SKU já está em uso.")
			err := Render(templates.ProductForm(form, false), r, w)

			if err != nil {
				http.Error(w, "Error Creating Product", http.StatusInternalServerError)
			}
			return
		}
		writeRegisterError(r, w, "Erro ao criar Produto. Tente novamente.")
		return
	}

	w.Header().Set(HXRedirect, "/produtos")
	w.WriteHeader(http.StatusOK)
}

func (p *ProductHandler) GetEditPage(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)
	var is_update = true

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	product, err := p.productStore.GetProduct(uint(id))

	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if product.TenantID != sess.TenantID {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	form := forms.New(nil)

	form.Set("ID", strconv.FormatUint(uint64(product.ID), 10))
	form.Set("name", product.Name)
	form.Set("sku", product.SKU)
	form.Set("uom", string(product.UOM))
	form.Set("description", product.FullDescription)
	form.Set("cost_price", strconv.FormatFloat(product.CostPrice, 'f', 2, 64))
	form.Set("weight", strconv.FormatFloat(product.Weight, 'f', 2, 64))
	form.Set("height", strconv.FormatFloat(product.HeightCm, 'f', 2, 64))
	form.Set("length", strconv.FormatFloat(product.LengthCm, 'f', 2, 64))
	form.Set("width", strconv.FormatFloat(product.WidthCm, 'f', 2, 64))

	form.Set("ean", product.EAN)
	form.Set("minimum_stock", strconv.FormatInt(int64(product.MinimumStock), 10))
	form.Set("current_stock", strconv.FormatInt(int64(product.CurrentStock), 10))

	Render(templates.ProductForm(form, is_update), r, w)
}

func (p *ProductHandler) GetProductPage(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)
	produtos, err := p.productStore.FindAllByUser(sess.TenantID)

	if err != nil {
		http.Error(w, "Error listing Product", http.StatusInternalServerError)
	}

	err = Render(templates.ProductsPage(produtos), r, w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
