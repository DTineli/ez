package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	m "github.com/DTineli/ez/internal/middleware"
	"github.com/DTineli/ez/internal/store"
	"github.com/go-chi/chi/v5"
)

// --- mocks ---

type mockProductStore struct {
	createProduct          func(*store.Product) error
	updateFields           func(id, tenantID uint, fields map[string]any) error
	getProduct             func(id uint) (*store.Product, error)
	findAllByUserFilters   func(id uint, f store.ProductFilters) (*store.FindResults[store.Product], error)
	findAllByUser          func(userID uint) ([]store.Product, error)
	createVariant          func(*store.Variant) error
	getVariant             func(id, tenantID uint) (*store.Variant, error)
	findVariantsByProduct  func(productID, tenantID uint) ([]store.Variant, error)
	updateVariantFields    func(id, tenantID uint, fields map[string]any) error
	deleteVariant          func(id, tenantID uint) error
	setVariantAttributes   func(variantID uint, ids []uint) error
	createAttribute        func(*store.Attribute) error
	getAttribute           func(id, tenantID uint) (*store.Attribute, error)
	findAttributesByTenant func(tenantID uint) ([]store.Attribute, error)
	deleteAttribute        func(id, tenantID uint) error
	createAttributeValue   func(*store.AttributeValue) error
	deleteAttributeValue   func(id, tenantID uint) error
}

func (s *mockProductStore) CreateProduct(p *store.Product) error {
	if s.createProduct != nil {
		return s.createProduct(p)
	}
	return nil
}
func (s *mockProductStore) UpdateFields(id, tenantID uint, fields map[string]any) error {
	if s.updateFields != nil {
		return s.updateFields(id, tenantID, fields)
	}
	return nil
}
func (s *mockProductStore) GetProduct(id uint) (*store.Product, error) {
	if s.getProduct != nil {
		return s.getProduct(id)
	}
	return &store.Product{ID: id, TenantID: 1, SKU: "PROD-01", Name: "Produto Teste"}, nil
}
func (s *mockProductStore) FindAllByUserWithFilters(id uint, f store.ProductFilters) (*store.FindResults[store.Product], error) {
	if s.findAllByUserFilters != nil {
		return s.findAllByUserFilters(id, f)
	}
	return &store.FindResults[store.Product]{Count: 0, Results: nil}, nil
}
func (s *mockProductStore) FindAllByUser(userID uint) ([]store.Product, error) {
	if s.findAllByUser != nil {
		return s.findAllByUser(userID)
	}
	return nil, nil
}
func (s *mockProductStore) CreateVariant(v *store.Variant) error {
	if s.createVariant != nil {
		return s.createVariant(v)
	}
	return nil
}
func (s *mockProductStore) GetVariant(id, tenantID uint) (*store.Variant, error) {
	if s.getVariant != nil {
		return s.getVariant(id, tenantID)
	}
	return &store.Variant{ID: id, SKU: "VAR-01", TenantID: tenantID}, nil
}
func (s *mockProductStore) FindVariantsByProduct(productID, tenantID uint) ([]store.Variant, error) {
	if s.findVariantsByProduct != nil {
		return s.findVariantsByProduct(productID, tenantID)
	}
	return nil, nil
}
func (s *mockProductStore) UpdateVariantFields(id, tenantID uint, fields map[string]any) error {
	if s.updateVariantFields != nil {
		return s.updateVariantFields(id, tenantID, fields)
	}
	return nil
}
func (s *mockProductStore) DeleteVariant(id, tenantID uint) error {
	if s.deleteVariant != nil {
		return s.deleteVariant(id, tenantID)
	}
	return nil
}
func (s *mockProductStore) SetVariantAttributes(variantID uint, ids []uint) error {
	if s.setVariantAttributes != nil {
		return s.setVariantAttributes(variantID, ids)
	}
	return nil
}
func (s *mockProductStore) CreateAttribute(a *store.Attribute) error {
	if s.createAttribute != nil {
		return s.createAttribute(a)
	}
	return nil
}
func (s *mockProductStore) GetAttribute(id, tenantID uint) (*store.Attribute, error) {
	if s.getAttribute != nil {
		return s.getAttribute(id, tenantID)
	}
	return nil, nil
}
func (s *mockProductStore) FindAttributesByTenant(tenantID uint) ([]store.Attribute, error) {
	if s.findAttributesByTenant != nil {
		return s.findAttributesByTenant(tenantID)
	}
	return nil, nil
}
func (s *mockProductStore) DeleteAttribute(id, tenantID uint) error {
	if s.deleteAttribute != nil {
		return s.deleteAttribute(id, tenantID)
	}
	return nil
}
func (s *mockProductStore) CreateAttributeValue(v *store.AttributeValue) error {
	if s.createAttributeValue != nil {
		return s.createAttributeValue(v)
	}
	return nil
}
func (s *mockProductStore) FindDefaultVariant(productID, tenantID uint) (*store.Variant, error) {
	return nil, nil
}

func (s *mockProductStore) DeleteAttributeValue(id, tenantID uint) error {
	if s.deleteAttributeValue != nil {
		return s.deleteAttributeValue(id, tenantID)
	}
	return nil
}

func (s *mockProductStore) FindOrCreateAttribute(name string, tenantID uint) (*store.Attribute, error) {
	return &store.Attribute{Name: name, TenantID: tenantID}, nil
}

func (s *mockProductStore) FindOrCreateAttributeValue(value string, attrID uint) (*store.AttributeValue, error) {
	return &store.AttributeValue{Value: value, AttributeID: attrID}, nil
}

type mockPriceTableStore struct{}

func (s *mockPriceTableStore) CreatePriceTable(p *store.PriceTable) error  { return nil }
func (s *mockPriceTableStore) FindAllByTenant(id uint) ([]store.PriceTable, error) {
	return nil, nil
}
func (s *mockPriceTableStore) GetOne(id, tenantID uint) (*store.PriceTable, error) {
	return nil, nil
}
func (s *mockPriceTableStore) HasContacts(priceTableID, tenantID uint) (bool, error) {
	return false, nil
}
func (s *mockPriceTableStore) Delete(id, tenantID uint) error { return nil }

// --- helpers ---

func newSession(tenantID uint) *store.Session {
	return &store.Session{
		TenantID:       tenantID,
		UserAccessType: store.AccessAdmin,
	}
}

func withSession(r *http.Request, sess *store.Session) *http.Request {
	ctx := context.WithValue(r.Context(), m.SessionInfoKey, sess)
	return r.WithContext(ctx)
}

func withChiParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func withChiParams(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func htmxRequest(r *http.Request) *http.Request {
	r.Header.Set("HX-Request", "true")
	return r
}

func newHandler() *ProductHandler {
	return NewProductHandler(&mockProductStore{}, &mockPriceTableStore{})
}

// --- testes ---

func TestGetProductForm(t *testing.T) {
	h := newHandler()

	r := httptest.NewRequest(http.MethodGet, "/admin/produtos/novo", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.GetProductForm(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestPostNewProduct_ValidacaoFalha(t *testing.T) {
	h := newHandler()

	body := url.Values{}
	// name e sku ausentes → validação falha
	r := httptest.NewRequest(http.MethodPost, "/admin/produtos", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.PostNewProduct(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
	if strings.Contains(w.Header().Get("HX-Trigger"), "error") {
		t.Error("não deve usar toast para erros de validação")
	}
}

func TestPostNewProduct_SKUDuplicado(t *testing.T) {
	ps := &mockProductStore{
		createProduct: func(p *store.Product) error {
			return errors.New("UNIQUE constraint failed: products.sku")
		},
	}
	h := NewProductHandler(ps, &mockPriceTableStore{})

	body := url.Values{
		"name": {"Produto Teste"},
		"sku":  {"PROD-01"},
		"uom":  {"UN"},
	}
	r := httptest.NewRequest(http.MethodPost, "/admin/produtos", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.PostNewProduct(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Este SKU já está em uso") {
		t.Error("esperado erro inline para SKU duplicado")
	}
}

func TestPostNewProduct_Sucesso(t *testing.T) {
	var criado *store.Product
	ps := &mockProductStore{
		createProduct: func(p *store.Product) error {
			p.ID = 42
			criado = p
			return nil
		},
	}
	h := NewProductHandler(ps, &mockPriceTableStore{})

	body := url.Values{
		"name": {"Produto Teste"},
		"sku":  {"PROD-01"},
		"uom":  {"UN"},
	}
	r := httptest.NewRequest(http.MethodPost, "/admin/produtos", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.PostNewProduct(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
	if criado == nil {
		t.Fatal("CreateProduct não foi chamado")
	}
	if criado.SKU != "PROD-01" {
		t.Errorf("SKU incorreto: %q", criado.SKU)
	}
	if criado.TenantID != 1 {
		t.Errorf("TenantID incorreto: %d", criado.TenantID)
	}
	if !strings.Contains(w.Header().Get("HX-Trigger"), "success") {
		t.Error("esperado toast de sucesso")
	}
}

func TestGetEditPage_ProdutoNaoEncontrado(t *testing.T) {
	ps := &mockProductStore{
		getProduct: func(id uint) (*store.Product, error) {
			return nil, errors.New("record not found")
		},
	}
	h := NewProductHandler(ps, &mockPriceTableStore{})

	r := httptest.NewRequest(http.MethodGet, "/admin/produtos/99", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParam(r, "id", "99")
	w := httptest.NewRecorder()

	h.GetEditPage(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("esperado 400, obteve %d", w.Code)
	}
}

func TestGetEditPage_TenantErrado(t *testing.T) {
	ps := &mockProductStore{
		getProduct: func(id uint) (*store.Product, error) {
			return &store.Product{ID: id, TenantID: 99}, nil // tenant diferente
		},
	}
	h := NewProductHandler(ps, &mockPriceTableStore{})

	r := httptest.NewRequest(http.MethodGet, "/admin/produtos/1", nil)
	r = htmxRequest(withSession(r, newSession(1))) // session tenant=1
	r = withChiParam(r, "id", "1")
	w := httptest.NewRecorder()

	h.GetEditPage(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("esperado 400, obteve %d", w.Code)
	}
}

func TestGetEditPage_Sucesso(t *testing.T) {
	ps := &mockProductStore{
		getProduct: func(id uint) (*store.Product, error) {
			return &store.Product{ID: id, TenantID: 1, SKU: "PROD-01", Name: "Teste", UOM: "UN"}, nil
		},
		findVariantsByProduct: func(productID, tenantID uint) ([]store.Variant, error) {
			return nil, nil
		},
	}
	h := NewProductHandler(ps, &mockPriceTableStore{})

	r := httptest.NewRequest(http.MethodGet, "/admin/produtos/1", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParam(r, "id", "1")
	w := httptest.NewRecorder()

	h.GetEditPage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestUpdateProduct_Sucesso(t *testing.T) {
	var camposAtualizados map[string]any
	ps := &mockProductStore{
		updateFields: func(id, tenantID uint, fields map[string]any) error {
			camposAtualizados = fields
			return nil
		},
		findVariantsByProduct: func(productID, tenantID uint) ([]store.Variant, error) {
			return nil, nil
		},
	}
	h := NewProductHandler(ps, &mockPriceTableStore{})

	body := url.Values{
		"name": {"Produto Atualizado"},
		"sku":  {"PROD-01"},
		"uom":  {"KG"},
	}
	r := httptest.NewRequest(http.MethodPost, "/admin/produtos/1", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParam(r, "id", "1")
	w := httptest.NewRecorder()

	h.UpdateProduct(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
	if camposAtualizados["name"] != "Produto Atualizado" {
		t.Errorf("campo name incorreto: %v", camposAtualizados["name"])
	}
	if !strings.Contains(w.Header().Get("HX-Trigger"), "success") {
		t.Error("esperado toast de sucesso")
	}
}

func TestGetProductPage_Sucesso(t *testing.T) {
	h := newHandler()

	r := httptest.NewRequest(http.MethodGet, "/admin/produtos?page=1", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.GetProductPage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestGetProductPage_ErroStore(t *testing.T) {
	ps := &mockProductStore{
		findAllByUserFilters: func(id uint, f store.ProductFilters) (*store.FindResults[store.Product], error) {
			return nil, errors.New("db error")
		},
	}
	h := NewProductHandler(ps, &mockPriceTableStore{})

	r := httptest.NewRequest(http.MethodGet, "/admin/produtos", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.GetProductPage(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("esperado 500, obteve %d", w.Code)
	}
}

func TestGetVariantForm_Sucesso(t *testing.T) {
	h := newHandler()

	r := httptest.NewRequest(http.MethodGet, "/admin/produtos/1/variants/form", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParam(r, "id", "1")
	w := httptest.NewRecorder()

	h.GetVariantForm(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestCancelVariantForm_Sucesso(t *testing.T) {
	h := newHandler()

	r := httptest.NewRequest(http.MethodGet, "/admin/produtos/1/variants/form/cancel", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParam(r, "id", "1")
	w := httptest.NewRecorder()

	h.CancelVariantForm(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestPostVariant_Sucesso(t *testing.T) {
	var criado *store.Variant
	ps := &mockProductStore{
		createVariant: func(v *store.Variant) error {
			v.ID = 10
			criado = v
			return nil
		},
		setVariantAttributes: func(variantID uint, ids []uint) error { return nil },
		findVariantsByProduct: func(productID, tenantID uint) ([]store.Variant, error) {
			return []store.Variant{}, nil
		},
	}
	h := NewProductHandler(ps, &mockPriceTableStore{})

	body := url.Values{
		"sku":                 {"VAR-01"},
		"cost_price":          {"19.90"},
		"current_stock":       {"5"},
		"minimum_stock":       {"1"},
		"attribute_value_ids": {"1"},
	}
	r := httptest.NewRequest(http.MethodPost, "/admin/produtos/1/variants", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParam(r, "id", "1")
	w := httptest.NewRecorder()

	h.PostVariant(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
	if criado == nil {
		t.Fatal("CreateVariant não foi chamado")
	}
	if criado.SKU != "VAR-01" {
		t.Errorf("SKU incorreto: %q", criado.SKU)
	}
	if criado.ProductID != 1 {
		t.Errorf("ProductID incorreto: %d", criado.ProductID)
	}
	if criado.CostPrice != 19.90 {
		t.Errorf("CostPrice incorreto: %v", criado.CostPrice)
	}
	if !strings.Contains(w.Header().Get("HX-Trigger"), "success") {
		t.Error("esperado toast de sucesso")
	}
}

func TestPostVariant_ErroCreate(t *testing.T) {
	ps := &mockProductStore{
		createVariant: func(v *store.Variant) error {
			return errors.New("db error")
		},
	}
	h := NewProductHandler(ps, &mockPriceTableStore{})

	body := url.Values{"sku": {"VAR-01"}}
	r := httptest.NewRequest(http.MethodPost, "/admin/produtos/1/variants", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParam(r, "id", "1")
	w := httptest.NewRecorder()

	h.PostVariant(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
	if !strings.Contains(w.Header().Get("HX-Trigger"), "error") {
		t.Error("esperado toast de erro")
	}
}

func TestUpdateVariant_Sucesso(t *testing.T) {
	var camposAtualizados map[string]any
	ps := &mockProductStore{
		updateVariantFields: func(id, tenantID uint, fields map[string]any) error {
			camposAtualizados = fields
			return nil
		},
	}
	h := NewProductHandler(ps, &mockPriceTableStore{})

	body := url.Values{
		"cost_price":    {"29.90"},
		"current_stock": {"10"},
		"minimum_stock": {"2"},
	}
	r := httptest.NewRequest(http.MethodPost, "/admin/produtos/1/variants/5", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParams(r, map[string]string{"id": "1", "variantID": "5"})
	w := httptest.NewRecorder()

	h.UpdateVariant(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
	if camposAtualizados["cost_price"] != 29.90 {
		t.Errorf("cost_price incorreto: %v", camposAtualizados["cost_price"])
	}
	if !strings.Contains(w.Header().Get("HX-Trigger"), "success") {
		t.Error("esperado toast de sucesso")
	}
}

func TestUpdateVariant_ErroStore(t *testing.T) {
	ps := &mockProductStore{
		updateVariantFields: func(id, tenantID uint, fields map[string]any) error {
			return errors.New("not found")
		},
	}
	h := NewProductHandler(ps, &mockPriceTableStore{})

	body := url.Values{"cost_price": {"10"}, "current_stock": {"1"}, "minimum_stock": {"0"}}
	r := httptest.NewRequest(http.MethodPost, "/admin/produtos/1/variants/5", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParams(r, map[string]string{"id": "1", "variantID": "5"})
	w := httptest.NewRecorder()

	h.UpdateVariant(w, r)

	if !strings.Contains(w.Header().Get("HX-Trigger"), "error") {
		t.Error("esperado toast de erro")
	}
}

func TestGetVariantRow_NaoEncontrado(t *testing.T) {
	ps := &mockProductStore{
		getVariant: func(id, tenantID uint) (*store.Variant, error) {
			return nil, errors.New("not found")
		},
	}
	h := NewProductHandler(ps, &mockPriceTableStore{})

	r := httptest.NewRequest(http.MethodGet, "/admin/produtos/1/variants/99", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParams(r, map[string]string{"id": "1", "variantID": "99"})
	w := httptest.NewRecorder()

	h.GetVariantRow(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("esperado 404, obteve %d", w.Code)
	}
}

func TestGetVariantRow_Sucesso(t *testing.T) {
	h := newHandler()

	r := httptest.NewRequest(http.MethodGet, "/admin/produtos/1/variants/5", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParams(r, map[string]string{"id": "1", "variantID": "5"})
	w := httptest.NewRecorder()

	h.GetVariantRow(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestGetEditVariantRow_Sucesso(t *testing.T) {
	h := newHandler()

	r := httptest.NewRequest(http.MethodGet, "/admin/produtos/1/variants/5/edit", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParams(r, map[string]string{"id": "1", "variantID": "5"})
	w := httptest.NewRecorder()

	h.GetEditVariantRow(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}
