package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DTineli/ez/internal/store"
)

// --- mocks ---

type mockOrderStore struct {
	confirmFromCart func(cartID, tenantID, contactID uint) (*store.Order, error)
	listByTenant    func(tenantID uint) ([]store.AdminOrderListItem, error)
	listByContact   func(tenantID, contactID uint) ([]store.ClientOrderListItem, error)
	getByID         func(id, tenantID uint) (*store.OrderDetail, error)
	create          func(tenantID, contactID uint, items []store.NewOrderItem) (*store.Order, error)
}

func (s *mockOrderStore) ConfirmFromCart(cartID, tenantID, contactID uint) (*store.Order, error) {
	if s.confirmFromCart != nil {
		return s.confirmFromCart(cartID, tenantID, contactID)
	}
	return &store.Order{}, nil
}
func (s *mockOrderStore) ListByTenant(tenantID uint) ([]store.AdminOrderListItem, error) {
	if s.listByTenant != nil {
		return s.listByTenant(tenantID)
	}
	return nil, nil
}
func (s *mockOrderStore) ListByContact(tenantID, contactID uint) ([]store.ClientOrderListItem, error) {
	if s.listByContact != nil {
		return s.listByContact(tenantID, contactID)
	}
	return nil, nil
}
func (s *mockOrderStore) GetByID(id, tenantID uint) (*store.OrderDetail, error) {
	if s.getByID != nil {
		return s.getByID(id, tenantID)
	}
	return &store.OrderDetail{}, nil
}
func (s *mockOrderStore) Create(tenantID, contactID uint, items []store.NewOrderItem) (*store.Order, error) {
	if s.create != nil {
		return s.create(tenantID, contactID, items)
	}
	return &store.Order{ID: 99}, nil
}

func newAdminOrderHandler(os *mockOrderStore, cs *mockContactStore, ps *mockProductStore) *AdminOrderHandler {
	if os == nil {
		os = &mockOrderStore{}
	}
	if cs == nil {
		cs = &mockContactStore{}
	}
	if ps == nil {
		ps = &mockProductStore{}
	}
	return NewAdminOrderHandler(os, cs, ps)
}

// --- testes ---

func TestAdminGetOrdersPage_Sucesso(t *testing.T) {
	h := newAdminOrderHandler(nil, nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/admin/pedidos", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.GetOrdersPage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestAdminGetOrdersPage_ErroStore(t *testing.T) {
	os := &mockOrderStore{
		listByTenant: func(tenantID uint) ([]store.AdminOrderListItem, error) {
			return nil, errors.New("db error")
		},
	}
	h := newAdminOrderHandler(os, nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/admin/pedidos", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.GetOrdersPage(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("esperado 500, obteve %d", w.Code)
	}
}

func TestAdminGetOrderPage_IDInvalido(t *testing.T) {
	h := newAdminOrderHandler(nil, nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/admin/pedidos/abc", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParam(r, "id", "abc")
	w := httptest.NewRecorder()

	h.GetOrderPage(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("esperado 400, obteve %d", w.Code)
	}
}

func TestAdminGetOrderPage_NaoEncontrado(t *testing.T) {
	os := &mockOrderStore{
		getByID: func(id, tenantID uint) (*store.OrderDetail, error) {
			return nil, errors.New("not found")
		},
	}
	h := newAdminOrderHandler(os, nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/admin/pedidos/99", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParam(r, "id", "99")
	w := httptest.NewRecorder()

	h.GetOrderPage(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("esperado 404, obteve %d", w.Code)
	}
}

func TestAdminGetOrderPage_Sucesso(t *testing.T) {
	h := newAdminOrderHandler(nil, nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/admin/pedidos/1", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParam(r, "id", "1")
	w := httptest.NewRecorder()

	h.GetOrderPage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestAdminGetNewOrderPage_Sucesso(t *testing.T) {
	h := newAdminOrderHandler(nil, nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/admin/pedidos/novo", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.GetNewOrderPage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestAdminGetNewOrderPage_ErroContatos(t *testing.T) {
	cs := &mockContactStore{
		findAll: func(tenantID uint, f store.ContactFilters) (*store.FindResults[store.Contact], error) {
			return nil, errors.New("db error")
		},
	}
	h := newAdminOrderHandler(nil, cs, nil)

	r := httptest.NewRequest(http.MethodGet, "/admin/pedidos/novo", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.GetNewOrderPage(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("esperado 500, obteve %d", w.Code)
	}
}

func TestSearchProductsForOrder_QueryCurta(t *testing.T) {
	h := newAdminOrderHandler(nil, nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/admin/pedidos/busca?q=a", nil) // < 2 chars
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.SearchProductsForOrder(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestSearchProductsForOrder_Sucesso(t *testing.T) {
	ps := &mockProductStore{
		findAllByUserFilters: func(id uint, f store.ProductFilters) (*store.FindResults[store.Product], error) {
			return &store.FindResults[store.Product]{
				Count:   1,
				Results: []store.Product{{ID: 1, Name: "Camiseta"}},
			}, nil
		},
	}
	h := newAdminOrderHandler(nil, nil, ps)

	r := httptest.NewRequest(http.MethodGet, "/admin/pedidos/busca?q=Cam", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.SearchProductsForOrder(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestAdminPostNewOrder_ContatoInvalido(t *testing.T) {
	h := newAdminOrderHandler(nil, nil, nil)

	body := url.Values{"contact_id": {"abc"}}
	r := httptest.NewRequest(http.MethodPost, "/admin/pedidos", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.PostNewOrder(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("esperado 400, obteve %d", w.Code)
	}
}

func TestAdminPostNewOrder_ItensFaltando(t *testing.T) {
	h := newAdminOrderHandler(nil, nil, nil)

	body := url.Values{"contact_id": {"1"}} // sem product_id[]
	r := httptest.NewRequest(http.MethodPost, "/admin/pedidos", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.PostNewOrder(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("esperado 400, obteve %d", w.Code)
	}
}

func TestAdminPostNewOrder_Sucesso(t *testing.T) {
	var pedidoCriado bool
	os := &mockOrderStore{
		create: func(tenantID, contactID uint, items []store.NewOrderItem) (*store.Order, error) {
			pedidoCriado = true
			return &store.Order{ID: 42}, nil
		},
	}
	h := newAdminOrderHandler(os, nil, nil)

	body := url.Values{
		"contact_id":   {"1"},
		"product_id[]": {"10"},
		"quantity[]":   {"2"},
		"unit_price[]": {"49.90"},
	}
	r := httptest.NewRequest(http.MethodPost, "/admin/pedidos", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.PostNewOrder(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
	if !pedidoCriado {
		t.Error("Create não foi chamado")
	}
	if !strings.Contains(w.Header().Get("HX-Redirect"), "/admin/pedidos/42") {
		t.Errorf("esperado redirect para pedido 42, obteve %q", w.Header().Get("HX-Redirect"))
	}
}
