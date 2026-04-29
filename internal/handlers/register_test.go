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
	"gorm.io/gorm"
)

func newRegisterHandler(us *mockUserStore, ts *mockTenantStore, is *mockInviteStore, cs *mockContactStore, ss *mockSessionStore) *RegisterHandler {
	if us == nil {
		us = &mockUserStore{}
	}
	if ts == nil {
		ts = &mockTenantStore{}
	}
	if is == nil {
		is = &mockInviteStore{}
	}
	if cs == nil {
		cs = &mockContactStore{}
	}
	if ss == nil {
		ss = &mockSessionStore{}
	}
	return NewRegisterHandler(us, ts, is, cs, ss)
}

var validToken = uuid.New()

// --- GetRegisterPage ---

func TestGetRegisterPage(t *testing.T) {
	h := newRegisterHandler(nil, nil, nil, nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/register", nil)
	w := httptest.NewRecorder()

	h.GetRegisterPage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

// --- GetRegisterClientPage ---

func TestGetRegisterClientPage_TokenAusente(t *testing.T) {
	h := newRegisterHandler(nil, nil, nil, nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/client/register", nil)
	w := httptest.NewRecorder()

	h.GetRegisterClientPage(w, r)

	if !strings.Contains(w.Body.String(), "Token de convite Invalido") {
		t.Error("esperado mensagem de token inválido")
	}
}

func TestGetRegisterClientPage_TokenInvalido(t *testing.T) {
	h := newRegisterHandler(nil, nil, nil, nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/client/register?token=nao-eh-uuid", nil)
	w := httptest.NewRecorder()

	h.GetRegisterClientPage(w, r)

	if !strings.Contains(w.Body.String(), "Token de convite Invalido") {
		t.Error("esperado mensagem de token inválido")
	}
}

func TestGetRegisterClientPage_TokenNaoEncontrado(t *testing.T) {
	is := &mockInviteStore{
		findByID: func(id uuid.UUID) (*store.Invite, error) {
			return nil, errors.New("not found")
		},
	}
	h := newRegisterHandler(nil, nil, is, nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/client/register?token="+validToken.String(), nil)
	w := httptest.NewRecorder()

	h.GetRegisterClientPage(w, r)

	if !strings.Contains(w.Body.String(), "Token de convite Invalido") {
		t.Error("esperado mensagem de token inválido")
	}
}

func TestGetRegisterClientPage_Sucesso(t *testing.T) {
	h := newRegisterHandler(nil, nil, nil, nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/client/register?token="+validToken.String(), nil)
	w := httptest.NewRecorder()

	h.GetRegisterClientPage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

// --- PostRegister (admin) ---

func TestPostRegister_NomeFaltando(t *testing.T) {
	h := newRegisterHandler(nil, nil, nil, nil, nil)

	body := url.Values{"name": {""}, "email": {"a@b.com"}, "slug": {"emp"}, "password": {"1234"}, "password_confirmation": {"1234"}}
	r := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.PostRegister(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestPostRegister_SenhasNaoBatem(t *testing.T) {
	h := newRegisterHandler(nil, nil, nil, nil, nil)

	body := url.Values{
		"name": {"Admin"}, "email": {"a@b.com"}, "slug": {"emp"},
		"password": {"1234"}, "password_confirmation": {"9999"},
	}
	r := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.PostRegister(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestPostRegister_SlugDuplicado(t *testing.T) {
	ts := &mockTenantStore{
		createTenant: func(t store.Tenant) (uint, error) {
			return 0, errors.New("UNIQUE constraint failed: tenants.slug")
		},
	}
	h := newRegisterHandler(nil, ts, nil, nil, nil)

	body := url.Values{
		"name": {"Admin"}, "email": {"a@b.com"}, "slug": {"existente"},
		"password": {"1234"}, "password_confirmation": {"1234"},
	}
	r := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Host = "existente.localhost"
	w := httptest.NewRecorder()

	h.PostRegister(w, r)

	if !strings.Contains(w.Body.String(), "Slug") {
		t.Error("esperado erro de slug duplicado")
	}
}

func TestPostRegister_Sucesso(t *testing.T) {
	h := newRegisterHandler(nil, nil, nil, nil, nil)

	body := url.Values{
		"name": {"Admin"}, "email": {"admin@emp.com"}, "slug": {"novaemp"},
		"password": {"1234"}, "password_confirmation": {"1234"},
	}
	r := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Host = "novaemp.localhost"
	w := httptest.NewRecorder()

	h.PostRegister(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
	if !strings.Contains(w.Header().Get(HXRedirect), "/admin") {
		t.Errorf("esperado redirect para /admin, obteve %q", w.Header().Get(HXRedirect))
	}
}

// --- PostRegisterClient ---

func TestPostRegisterClient_TokenInvalido(t *testing.T) {
	h := newRegisterHandler(nil, nil, nil, nil, nil)

	body := url.Values{"invite_token": {"nao-eh-uuid"}}
	r := httptest.NewRequest(http.MethodPost, "/client/register", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.PostRegisterClient(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestPostRegisterClient_SenhasCurtaDemais(t *testing.T) {
	h := newRegisterHandler(nil, nil, nil, nil, nil)

	body := url.Values{
		"invite_token": {validToken.String()},
		"name": {"João"}, "email": {"j@t.com"},
		"password": {"12"}, "password_confirmation": {"12"},
		"contact_id": {"1"}, "tenant_id": {"1"},
	}
	r := httptest.NewRequest(http.MethodPost, "/client/register", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.PostRegisterClient(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestPostRegisterClient_UsuarioNovo_Sucesso(t *testing.T) {
	us := &mockUserStore{
		getUserByPhone: func(phone string) (*store.User, error) {
			return nil, gorm.ErrRecordNotFound // usuário novo
		},
		createUser: func(u *store.User) error {
			u.ID = 20
			return nil
		},
	}
	cs := &mockContactStore{
		getOne: func(id uint) (*store.Contact, error) {
			return &store.Contact{ID: id, TenantID: 1, PriceTableID: 2}, nil
		},
	}
	h := newRegisterHandler(us, nil, nil, cs, nil)

	body := url.Values{
		"invite_token": {validToken.String()},
		"name": {"João"}, "email": {"j@test.com"},
		"password": {"senha123"}, "password_confirmation": {"senha123"},
		"phone": {"11999999999"}, "contact_id": {"5"}, "tenant_id": {"1"},
	}
	r := httptest.NewRequest(http.MethodPost, "/client/register", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.PostRegisterClient(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
	if w.Header().Get(HXRedirect) != "/client/items" {
		t.Errorf("esperado redirect para /client/items, obteve %q", w.Header().Get(HXRedirect))
	}
}

func TestPostRegisterClient_UsuarioExistente_Sucesso(t *testing.T) {
	// Usuário já existe (telefone encontrado) → só vincula o contato
	us := &mockUserStore{
		getUserByPhone: func(phone string) (*store.User, error) {
			return &store.User{ID: 7, Phone: phone}, nil
		},
	}
	cs := &mockContactStore{
		getOne: func(id uint) (*store.Contact, error) {
			return &store.Contact{ID: id, TenantID: 1, PriceTableID: 2}, nil
		},
	}
	h := newRegisterHandler(us, nil, nil, cs, nil)

	body := url.Values{
		"invite_token": {validToken.String()},
		"name": {"João"}, "email": {"j@test.com"},
		"password": {"senha123"}, "password_confirmation": {"senha123"},
		"phone": {"11999999999"}, "contact_id": {"5"}, "tenant_id": {"1"},
	}
	r := httptest.NewRequest(http.MethodPost, "/client/register", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.PostRegisterClient(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
	if w.Header().Get(HXRedirect) != "/client/items" {
		t.Errorf("esperado redirect para /client/items, obteve %q", w.Header().Get(HXRedirect))
	}
}
