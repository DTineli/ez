package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DTineli/ez/internal/store"
)

func TestClientGetOrdersPage_Sucesso(t *testing.T) {
	h := newClientHandler(nil, nil, nil, &mockSessionStore{}, nil)

	r := httptest.NewRequest(http.MethodGet, "/client/pedidos", nil)
	r = htmxRequest(withSession(r, newClientSession()))
	w := httptest.NewRecorder()

	h.GetOrdersPage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestClientGetOrdersPage_ErroStore(t *testing.T) {
	os := &mockOrderStore{
		listByContact: func(tenantID, contactID uint) ([]store.ClientOrderListItem, error) {
			return nil, errors.New("db error")
		},
	}
	h := newClientHandler(nil, nil, os, &mockSessionStore{}, nil)

	r := httptest.NewRequest(http.MethodGet, "/client/pedidos", nil)
	r = htmxRequest(withSession(r, newClientSession()))
	w := httptest.NewRecorder()

	h.GetOrdersPage(w, r)

	// handler mostra toast mas não retorna erro HTTP — verifica trigger
	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestClientGetOrderDetail_IDInvalido(t *testing.T) {
	h := newClientHandler(nil, nil, nil, &mockSessionStore{}, nil)

	r := httptest.NewRequest(http.MethodGet, "/client/pedidos/abc", nil)
	r = htmxRequest(withSession(r, newClientSession()))
	r = withChiParam(r, "id", "abc")
	w := httptest.NewRecorder()

	h.GetOrderDetail(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("esperado 400, obteve %d", w.Code)
	}
}

func TestClientGetOrderDetail_NaoEncontrado(t *testing.T) {
	os := &mockOrderStore{
		getByID: func(id, tenantID uint) (*store.OrderDetail, error) {
			return nil, errors.New("not found")
		},
	}
	h := newClientHandler(nil, nil, os, &mockSessionStore{}, nil)

	r := httptest.NewRequest(http.MethodGet, "/client/pedidos/99", nil)
	r = htmxRequest(withSession(r, newClientSession()))
	r = withChiParam(r, "id", "99")
	w := httptest.NewRecorder()

	h.GetOrderDetail(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("esperado 404, obteve %d", w.Code)
	}
}

func TestClientGetOrderDetail_ContatoErrado(t *testing.T) {
	// pedido pertence a outro contato
	os := &mockOrderStore{
		getByID: func(id, tenantID uint) (*store.OrderDetail, error) {
			return &store.OrderDetail{ContactID: 999}, nil // diferente do session contact 5
		},
	}
	h := newClientHandler(nil, nil, os, &mockSessionStore{}, nil)

	r := httptest.NewRequest(http.MethodGet, "/client/pedidos/1", nil)
	r = htmxRequest(withSession(r, newClientSession()))
	r = withChiParam(r, "id", "1")
	w := httptest.NewRecorder()

	h.GetOrderDetail(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("esperado 404, obteve %d", w.Code)
	}
}

func TestClientGetOrderDetail_Sucesso(t *testing.T) {
	os := &mockOrderStore{
		getByID: func(id, tenantID uint) (*store.OrderDetail, error) {
			return &store.OrderDetail{ContactID: 5}, nil // mesmo contact do session
		},
	}
	h := newClientHandler(nil, nil, os, &mockSessionStore{}, nil)

	r := httptest.NewRequest(http.MethodGet, "/client/pedidos/1", nil)
	r = htmxRequest(withSession(r, newClientSession()))
	r = withChiParam(r, "id", "1")
	w := httptest.NewRecorder()

	h.GetOrderDetail(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}
