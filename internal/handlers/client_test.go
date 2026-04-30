package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DTineli/ez/internal/store"
	"gorm.io/gorm"
)

// --- mocks ---

type mockCartStore struct {
	findOpenByID      func(id, tenantID, contactID uint) (*store.Cart, error)
	findOpenByContact func(tenantID, contactID uint) (*store.Cart, error)
	create            func(*store.Cart) error
	addOrIncrementItem func(cartID, productID, variantID uint, quantity int, unitPrice float64) error
	countItems        func(cartID uint) (int64, error)
	listCheckoutItems func(cartID, tenantID uint) ([]store.CartCheckoutItem, error)
	removeItem        func(cartID, productID uint) error
	updateItemQty     func(cartID, productID uint, quantity int) error
}

func (s *mockCartStore) FindOpenByID(id, tenantID, contactID uint) (*store.Cart, error) {
	if s.findOpenByID != nil {
		return s.findOpenByID(id, tenantID, contactID)
	}
	return nil, gorm.ErrRecordNotFound
}
func (s *mockCartStore) FindOpenByContact(tenantID, contactID uint) (*store.Cart, error) {
	if s.findOpenByContact != nil {
		return s.findOpenByContact(tenantID, contactID)
	}
	return nil, gorm.ErrRecordNotFound
}
func (s *mockCartStore) Create(c *store.Cart) error {
	if s.create != nil {
		return s.create(c)
	}
	c.ID = 1
	return nil
}
func (s *mockCartStore) AddOrIncrementItem(cartID, productID, variantID uint, quantity int, unitPrice float64) error {
	if s.addOrIncrementItem != nil {
		return s.addOrIncrementItem(cartID, productID, variantID, quantity, unitPrice)
	}
	return nil
}
func (s *mockCartStore) CountItems(cartID uint) (int64, error) {
	if s.countItems != nil {
		return s.countItems(cartID)
	}
	return 0, nil
}
func (s *mockCartStore) ListCheckoutItems(cartID, tenantID uint) ([]store.CartCheckoutItem, error) {
	if s.listCheckoutItems != nil {
		return s.listCheckoutItems(cartID, tenantID)
	}
	return nil, nil
}
func (s *mockCartStore) RemoveItem(cartID, productID uint) error {
	if s.removeItem != nil {
		return s.removeItem(cartID, productID)
	}
	return nil
}
func (s *mockCartStore) UpdateItemQty(cartID, productID uint, quantity int) error {
	if s.updateItemQty != nil {
		return s.updateItemQty(cartID, productID, quantity)
	}
	return nil
}

func newClientHandler(ps *mockProductStore, cs *mockCartStore, os *mockOrderStore, ss *mockSessionStore, pts *mockPriceTableStoreExt) *ClientHandler {
	if ps == nil {
		ps = &mockProductStore{}
	}
	if cs == nil {
		cs = &mockCartStore{}
	}
	if os == nil {
		os = &mockOrderStore{}
	}
	if ss == nil {
		ss = &mockSessionStore{}
	}
	if pts == nil {
		pts = &mockPriceTableStoreExt{}
	}
	return NewClientHandler(ps, cs, os, ss, pts)
}

func newClientSession() *store.Session {
	return &store.Session{
		TenantID:       1,
		UserAccessType: store.AccessCustomer,
		ContactInfo:    &store.ContactInfo{ID: 5, PriceTable: 2},
	}
}

func newClientSessionWithCart(cartID uint) *store.Session {
	s := newClientSession()
	s.CartID = cartID
	return s
}

// --- GetItemsPage ---

func TestGetItemsPage_Sucesso(t *testing.T) {
	pts := &mockPriceTableStoreExt{
		getOne: func(id, tenantID uint) (*store.PriceTable, error) {
			return &store.PriceTable{ID: id, Percentage: 10}, nil
		},
	}
	h := newClientHandler(nil, nil, nil, nil, pts)

	r := httptest.NewRequest(http.MethodGet, "/client/items", nil)
	r = htmxRequest(withSession(r, newClientSession()))
	w := httptest.NewRecorder()

	h.GetItemsPage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestGetItemsPage_TabelaPrecoNaoEncontrada(t *testing.T) {
	pts := &mockPriceTableStoreExt{
		getOne: func(id, tenantID uint) (*store.PriceTable, error) {
			return nil, errors.New("not found")
		},
	}
	h := newClientHandler(nil, nil, nil, nil, pts)

	r := httptest.NewRequest(http.MethodGet, "/client/items", nil)
	r = htmxRequest(withSession(r, newClientSession()))
	w := httptest.NewRecorder()

	h.GetItemsPage(w, r)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("esperado 422, obteve %d", w.Code)
	}
}

// --- PostAddToCart ---

func TestPostAddToCart_ProdutoIDInvalido(t *testing.T) {
	h := newClientHandler(nil, nil, nil, nil, nil)

	body := url.Values{"product_id": {"abc"}, "qty": {"1"}}
	r := httptest.NewRequest(http.MethodPost, "/client/cart/items", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newClientSession()))
	w := httptest.NewRecorder()

	h.PostAddToCart(w, r)

	if !strings.Contains(w.Header().Get("HX-Trigger"), "error") {
		t.Error("esperado toast de erro para product_id inválido")
	}
}

func TestPostAddToCart_QuantidadeInvalida(t *testing.T) {
	h := newClientHandler(nil, nil, nil, nil, nil)

	body := url.Values{"product_id": {"1"}, "qty": {"0"}}
	r := httptest.NewRequest(http.MethodPost, "/client/cart/items", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newClientSession()))
	w := httptest.NewRecorder()

	h.PostAddToCart(w, r)

	if !strings.Contains(w.Header().Get("HX-Trigger"), "error") {
		t.Error("esperado toast de erro para quantidade inválida")
	}
}

func TestPostAddToCart_ProdutoNaoEncontrado(t *testing.T) {
	ps := &mockProductStore{
		getProduct: func(id uint) (*store.Product, error) {
			return nil, errors.New("not found")
		},
	}
	h := newClientHandler(ps, nil, nil, nil, nil)

	body := url.Values{"product_id": {"99"}, "qty": {"1"}}
	r := httptest.NewRequest(http.MethodPost, "/client/cart/items", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newClientSession()))
	w := httptest.NewRecorder()

	h.PostAddToCart(w, r)

	if !strings.Contains(w.Header().Get("HX-Trigger"), "error") {
		t.Error("esperado toast de erro para produto não encontrado")
	}
}

func TestPostAddToCart_Sucesso(t *testing.T) {
	ps := &mockProductStore{
		getProduct: func(id uint) (*store.Product, error) {
			return &store.Product{
				ID:       id,
				TenantID: 1,
			}, nil
		},
		getVariant: func(id, tenantID uint) (*store.Variant, error) {
			return &store.Variant{ID: id, TenantID: tenantID, ProductID: 1, CostPrice: 50.0}, nil
		},
	}
	pts := &mockPriceTableStoreExt{
		getOne: func(id, tenantID uint) (*store.PriceTable, error) {
			return &store.PriceTable{ID: id, Percentage: 10}, nil
		},
	}
	var itemAdicionado bool
	cs := &mockCartStore{
		addOrIncrementItem: func(cartID, productID, variantID uint, quantity int, unitPrice float64) error {
			itemAdicionado = true
			return nil
		},
	}
	h := newClientHandler(ps, cs, nil, &mockSessionStore{}, pts)

	body := url.Values{"product_id": {"1"}, "variant_id": {"5"}, "qty": {"2"}}
	r := httptest.NewRequest(http.MethodPost, "/client/cart/items", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newClientSession()))
	w := httptest.NewRecorder()

	h.PostAddToCart(w, r)

	if !itemAdicionado {
		t.Error("AddOrIncrementItem não foi chamado")
	}
	if !strings.Contains(w.Header().Get("HX-Trigger"), "success") {
		t.Error("esperado trigger de sucesso")
	}
}

// --- DeleteCartItem ---

func TestDeleteCartItem_IDInvalido(t *testing.T) {
	h := newClientHandler(nil, nil, nil, nil, nil)

	r := httptest.NewRequest(http.MethodDelete, "/client/cart/items/abc", nil)
	r = htmxRequest(withSession(r, newClientSession()))
	r = withChiParam(r, "productID", "abc")
	w := httptest.NewRecorder()

	h.DeleteCartItem(w, r)

	if !strings.Contains(w.Header().Get("HX-Trigger"), "error") {
		t.Error("esperado toast de erro para ID inválido")
	}
}

func TestDeleteCartItem_Sucesso(t *testing.T) {
	removido := false
	cs := &mockCartStore{
		findOpenByContact: func(tenantID, contactID uint) (*store.Cart, error) {
			return &store.Cart{ID: 1}, nil
		},
		removeItem: func(cartID, productID uint) error {
			removido = true
			return nil
		},
	}
	h := newClientHandler(nil, cs, nil, &mockSessionStore{}, nil)

	r := httptest.NewRequest(http.MethodDelete, "/client/cart/items/10", nil)
	r = htmxRequest(withSession(r, newClientSession()))
	r = withChiParam(r, "productID", "10")
	w := httptest.NewRecorder()

	h.DeleteCartItem(w, r)

	if !removido {
		t.Error("RemoveItem não foi chamado")
	}
	if w.Header().Get(HXRedirect) != "/client/confirmacao" {
		t.Errorf("esperado redirect para /client/confirmacao, obteve %q", w.Header().Get(HXRedirect))
	}
}

// --- PatchCartItemQty ---

func TestPatchCartItemQty_QuantidadeInvalida(t *testing.T) {
	h := newClientHandler(nil, nil, nil, nil, nil)

	body := url.Values{"qty": {"0"}}
	r := httptest.NewRequest(http.MethodPatch, "/client/cart/items/1", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newClientSession()))
	r = withChiParam(r, "productID", "1")
	w := httptest.NewRecorder()

	h.PatchCartItemQty(w, r)

	if !strings.Contains(w.Header().Get("HX-Trigger"), "error") {
		t.Error("esperado toast de erro para quantidade inválida")
	}
}

func TestPatchCartItemQty_Sucesso(t *testing.T) {
	atualizado := false
	cs := &mockCartStore{
		findOpenByContact: func(tenantID, contactID uint) (*store.Cart, error) {
			return &store.Cart{ID: 1}, nil
		},
		updateItemQty: func(cartID, productID uint, quantity int) error {
			atualizado = true
			return nil
		},
	}
	h := newClientHandler(nil, cs, nil, &mockSessionStore{}, nil)

	body := url.Values{"qty": {"3"}}
	r := httptest.NewRequest(http.MethodPatch, "/client/cart/items/1", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newClientSession()))
	r = withChiParam(r, "productID", "1")
	w := httptest.NewRecorder()

	h.PatchCartItemQty(w, r)

	if !atualizado {
		t.Error("UpdateItemQty não foi chamado")
	}
	if w.Header().Get(HXRedirect) != "/client/confirmacao" {
		t.Errorf("esperado redirect para /client/confirmacao, obteve %q", w.Header().Get(HXRedirect))
	}
}

// --- PostConfirmOrder ---

func TestPostConfirmOrder_CarrinhoVazio(t *testing.T) {
	cs := &mockCartStore{
		findOpenByContact: func(tenantID, contactID uint) (*store.Cart, error) {
			return nil, errors.New("not found")
		},
	}
	h := newClientHandler(nil, cs, nil, &mockSessionStore{}, nil)

	r := httptest.NewRequest(http.MethodPost, "/client/confirmacao", nil)
	r = htmxRequest(withSession(r, newClientSession()))
	w := httptest.NewRecorder()

	h.PostConfirmOrder(w, r)

	if !strings.Contains(w.Header().Get("HX-Trigger"), "error") {
		t.Error("esperado toast de erro para carrinho vazio")
	}
}

func TestPostConfirmOrder_Sucesso(t *testing.T) {
	confirmado := false
	cs := &mockCartStore{
		findOpenByContact: func(tenantID, contactID uint) (*store.Cart, error) {
			return &store.Cart{ID: 1}, nil
		},
	}
	os := &mockOrderStore{
		confirmFromCart: func(cartID, tenantID, contactID uint) (*store.Order, error) {
			confirmado = true
			return &store.Order{ID: 10}, nil
		},
	}
	h := newClientHandler(nil, cs, os, &mockSessionStore{}, nil)

	r := httptest.NewRequest(http.MethodPost, "/client/confirmacao", nil)
	r = htmxRequest(withSession(r, newClientSession()))
	w := httptest.NewRecorder()

	h.PostConfirmOrder(w, r)

	if !confirmado {
		t.Error("ConfirmFromCart não foi chamado")
	}
	if w.Header().Get(HXRedirect) != "/client/items" {
		t.Errorf("esperado redirect para /client/items, obteve %q", w.Header().Get(HXRedirect))
	}
}

// --- GetCheckoutPage ---

func TestGetCheckoutPage_Sucesso(t *testing.T) {
	h := newClientHandler(nil, nil, nil, &mockSessionStore{}, nil)

	r := httptest.NewRequest(http.MethodGet, "/client/confirmacao", nil)
	r = htmxRequest(withSession(r, newClientSession()))
	w := httptest.NewRecorder()

	h.GetCheckoutPage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}
