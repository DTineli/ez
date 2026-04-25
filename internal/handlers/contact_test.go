package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DTineli/ez/internal/store"
	"github.com/google/uuid"
)

// --- mocks ---

type mockContactStore struct {
	createContact func(*store.Contact) error
	findAll       func(uint, store.ContactFilters) (*store.FindResults[store.Contact], error)
	getOne        func(uint) (*store.Contact, error)
	updateById    func(id, tenantID uint, fields map[string]any) error
}

func (s *mockContactStore) CreateContact(c *store.Contact) error {
	if s.createContact != nil {
		return s.createContact(c)
	}
	return nil
}
func (s *mockContactStore) FindAll(tenantID uint, f store.ContactFilters) (*store.FindResults[store.Contact], error) {
	if s.findAll != nil {
		return s.findAll(tenantID, f)
	}
	return &store.FindResults[store.Contact]{}, nil
}
func (s *mockContactStore) GetOne(id uint) (*store.Contact, error) {
	if s.getOne != nil {
		return s.getOne(id)
	}
	return &store.Contact{ID: id, TenantID: 1, Name: "Contato Teste", TradeName: "Trade"}, nil
}
func (s *mockContactStore) UpdateById(id, tenantID uint, fields map[string]any) error {
	if s.updateById != nil {
		return s.updateById(id, tenantID, fields)
	}
	return nil
}

type mockInviteStore struct {
	create    func(*store.Invite) error
	findByID  func(uuid.UUID) (*store.Invite, error)
	deleteByID func(uuid.UUID) error
}

func (s *mockInviteStore) Create(i *store.Invite) error {
	if s.create != nil {
		return s.create(i)
	}
	return nil
}
func (s *mockInviteStore) FindByID(id uuid.UUID) (*store.Invite, error) {
	if s.findByID != nil {
		return s.findByID(id)
	}
	return &store.Invite{ID: id}, nil
}
func (s *mockInviteStore) DeleteByID(id uuid.UUID) error {
	if s.deleteByID != nil {
		return s.deleteByID(id)
	}
	return nil
}

func newContactHandler(cs *mockContactStore, is *mockInviteStore, pts *mockPriceTableStore) *ContactHandler {
	if cs == nil {
		cs = &mockContactStore{}
	}
	if is == nil {
		is = &mockInviteStore{}
	}
	if pts == nil {
		pts = &mockPriceTableStore{}
	}
	return NewContactHandler(NewContactHandlerParams{
		Contact:    cs,
		Invite:     is,
		PriceTable: pts,
	})
}

var validContactBody = url.Values{
	"name":           {"Empresa Teste"},
	"trade_name":     {"Trade Teste"},
	"phone":          {"11999999999"},
	"contact_type":   {"customer"},
	"document_type":  {"cnpj"},
	"price_table_id": {"1"},
}

// --- testes ---

func TestGetContactsForm(t *testing.T) {
	h := newContactHandler(nil, nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/admin/contatos/novo", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.GetContactsForm(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestPostNewContact_ValidacaoFalha(t *testing.T) {
	h := newContactHandler(nil, nil, nil)

	body := url.Values{} // campos obrigatórios ausentes
	r := httptest.NewRequest(http.MethodPost, "/admin/contatos", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.PostNewContact(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
	if strings.Contains(w.Header().Get("HX-Trigger"), "error") {
		t.Error("não deve usar toast para erros de validação")
	}
}

func TestPostNewContact_Sucesso(t *testing.T) {
	var criado *store.Contact
	cs := &mockContactStore{
		createContact: func(c *store.Contact) error {
			c.ID = 10
			criado = c
			return nil
		},
	}
	h := newContactHandler(cs, nil, nil)

	r := httptest.NewRequest(http.MethodPost, "/admin/contatos", strings.NewReader(validContactBody.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.PostNewContact(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
	if criado == nil {
		t.Fatal("CreateContact não foi chamado")
	}
	if criado.TenantID != 1 {
		t.Errorf("TenantID incorreto: %d", criado.TenantID)
	}
	if !strings.Contains(w.Header().Get("HX-Trigger"), "success") {
		t.Error("esperado toast de sucesso")
	}
}

func TestPostNewContact_ErroStore(t *testing.T) {
	cs := &mockContactStore{
		createContact: func(c *store.Contact) error {
			return errors.New("db error")
		},
	}
	h := newContactHandler(cs, nil, nil)

	r := httptest.NewRequest(http.MethodPost, "/admin/contatos", strings.NewReader(validContactBody.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.PostNewContact(w, r)

	if !strings.Contains(w.Body.String(), "Erro ao cadastrar contato") {
		t.Error("esperado erro inline no form")
	}
}

func TestGetContactEditPage_NaoEncontrado(t *testing.T) {
	cs := &mockContactStore{
		getOne: func(id uint) (*store.Contact, error) {
			return nil, errors.New("not found")
		},
	}
	h := newContactHandler(cs, nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/admin/contatos/99", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParam(r, "id", "99")
	w := httptest.NewRecorder()

	h.GetEditPage(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("esperado 400, obteve %d", w.Code)
	}
}

func TestGetContactEditPage_TenantErrado(t *testing.T) {
	cs := &mockContactStore{
		getOne: func(id uint) (*store.Contact, error) {
			return &store.Contact{ID: id, TenantID: 99}, nil
		},
	}
	h := newContactHandler(cs, nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/admin/contatos/1", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParam(r, "id", "1")
	w := httptest.NewRecorder()

	h.GetEditPage(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("esperado 400, obteve %d", w.Code)
	}
}

func TestGetContactEditPage_Sucesso(t *testing.T) {
	h := newContactHandler(nil, nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/admin/contatos/1", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParam(r, "id", "1")
	w := httptest.NewRecorder()

	h.GetEditPage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestUpdateContact_ValidacaoFalha(t *testing.T) {
	h := newContactHandler(nil, nil, nil)

	body := url.Values{} // campos obrigatórios ausentes
	r := httptest.NewRequest(http.MethodPost, "/admin/contatos/1", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParam(r, "id", "1")
	w := httptest.NewRecorder()

	h.Update(w, r)

	if strings.Contains(w.Header().Get("HX-Trigger"), "error") {
		t.Error("não deve usar toast para erros de validação")
	}
}

func TestUpdateContact_Sucesso(t *testing.T) {
	var camposAtualizados map[string]any
	cs := &mockContactStore{
		updateById: func(id, tenantID uint, fields map[string]any) error {
			camposAtualizados = fields
			return nil
		},
	}
	h := newContactHandler(cs, nil, nil)

	r := httptest.NewRequest(http.MethodPost, "/admin/contatos/1", strings.NewReader(validContactBody.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParam(r, "id", "1")
	w := httptest.NewRecorder()

	h.Update(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
	if camposAtualizados["name"] != "Empresa Teste" {
		t.Errorf("campo name incorreto: %v", camposAtualizados["name"])
	}
	if !strings.Contains(w.Header().Get("HX-Trigger"), "success") {
		t.Error("esperado toast de sucesso")
	}
}

func TestGetContactsPage_Sucesso(t *testing.T) {
	h := newContactHandler(nil, nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/admin/contatos", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.GetContactsPage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestGetContactsPage_ErroStore(t *testing.T) {
	cs := &mockContactStore{
		findAll: func(tenantID uint, f store.ContactFilters) (*store.FindResults[store.Contact], error) {
			return nil, errors.New("db error")
		},
	}
	h := newContactHandler(cs, nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/admin/contatos", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.GetContactsPage(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("esperado 500, obteve %d", w.Code)
	}
}

func TestCreateLink_ContatoNaoEncontrado(t *testing.T) {
	cs := &mockContactStore{
		getOne: func(id uint) (*store.Contact, error) {
			return nil, errors.New("not found")
		},
	}
	h := newContactHandler(cs, nil, nil)

	r := httptest.NewRequest(http.MethodPost, "/admin/contatos/1/link", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParam(r, "id", "1")
	w := httptest.NewRecorder()

	h.CreateLink(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("esperado 400, obteve %d", w.Code)
	}
}

func TestCreateLink_Sucesso(t *testing.T) {
	cs := &mockContactStore{
		getOne: func(id uint) (*store.Contact, error) {
			return &store.Contact{ID: id, TenantID: 1, Document: "12345678", Phone: "11999999999"}, nil
		},
	}
	h := newContactHandler(cs, nil, nil)

	r := httptest.NewRequest(http.MethodPost, "/admin/contatos/1/link", nil)
	r.Host = "empresa.localhost"
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParam(r, "id", "1")
	w := httptest.NewRecorder()

	h.CreateLink(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}
