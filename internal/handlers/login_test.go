package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DTineli/ez/internal/store"
	"golang.org/x/crypto/bcrypt"
)

// --- mocks ---

type mockUserStore struct {
	createUser     func(*store.User) error
	getUser        func(email string) (*store.User, error)
	getUserByPhone func(phone string) (*store.User, error)
}

func (s *mockUserStore) CreateUser(u *store.User) error {
	if s.createUser != nil {
		return s.createUser(u)
	}
	return nil
}
func (s *mockUserStore) GetUser(email string) (*store.User, error) {
	if s.getUser != nil {
		return s.getUser(email)
	}
	return nil, errors.New("not found")
}
func (s *mockUserStore) GetUserByPhone(phone string) (*store.User, error) {
	if s.getUserByPhone != nil {
		return s.getUserByPhone(phone)
	}
	return nil, errors.New("not found")
}

type mockTenantStore struct {
	createTenant    func(store.Tenant) (uint, error)
	getTenantByID   func(id uint) (*store.Tenant, error)
	getTenantBySlug func(slug string) (*store.Tenant, error)
}

func (s *mockTenantStore) CreateTenant(t store.Tenant) (uint, error) {
	if s.createTenant != nil {
		return s.createTenant(t)
	}
	return 1, nil
}

func (s *mockTenantStore) GetTenantByID(id uint) (*store.Tenant, error) {
	if s.getTenantByID != nil {
		return s.getTenantByID(id)
	}
	return &store.Tenant{ID: id, Slug: "empresa"}, nil
}
func (s *mockTenantStore) GetTenantBySlug(slug string) (*store.Tenant, error) {
	if s.getTenantBySlug != nil {
		return s.getTenantBySlug(slug)
	}
	return &store.Tenant{ID: 1, Slug: slug}, nil
}

type mockSessionStore struct {
	createSession  func(*http.Request, http.ResponseWriter, store.Session) error
	deleteSession  func(*http.Request, http.ResponseWriter) error
	getSessionInfo func(*http.Request) (*store.Session, error)
	setCartID      func(*http.Request, http.ResponseWriter, uint) error
}

func (s *mockSessionStore) CreateSession(
	r *http.Request,
	w http.ResponseWriter,
	sess store.Session,
) error {
	if s.createSession != nil {
		return s.createSession(r, w, sess)
	}
	return nil
}

func (s *mockSessionStore) DeleteSession(
	r *http.Request,
	w http.ResponseWriter,
) error {
	if s.deleteSession != nil {
		return s.deleteSession(r, w)
	}
	return nil
}

func (s *mockSessionStore) GetSessionInfo(
	r *http.Request,
) (*store.Session, error) {
	if s.getSessionInfo != nil {
		return s.getSessionInfo(r)
	}
	return nil, errors.New("no session")
}

func (s *mockSessionStore) SetCartID(
	r *http.Request,
	w http.ResponseWriter,
	id uint,
) error {
	if s.setCartID != nil {
		return s.setCartID(r, w, id)
	}
	return nil
}

func newLoginHandler(
	us *mockUserStore,
	ts *mockTenantStore,
	ss *mockSessionStore,
) *LoginHandler {
	if us == nil {
		us = &mockUserStore{}
	}
	if ts == nil {
		ts = &mockTenantStore{}
	}
	if ss == nil {
		ss = &mockSessionStore{}
	}
	return NewLoginHandler(LoginHandlerParams{
		UserStore:    us,
		SessionStore: ss,
		TenantStore:  ts,
	})
}

func hashedPassword(t *testing.T, plain string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	return string(h)
}

// --- testes ---

func TestGetAdminLoginPage(t *testing.T) {
	h := newLoginHandler(nil, nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/admin/login", nil)
	w := httptest.NewRecorder()

	h.GetAdminLoginPage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestGetClientLoginPage(t *testing.T) {
	h := newLoginHandler(nil, nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/client/login", nil)
	w := httptest.NewRecorder()

	h.GetClientLoginPage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestAdminLogin_EmailVazio(t *testing.T) {
	h := newLoginHandler(nil, nil, nil)

	body := url.Values{"email": {""}, "password": {"senha123"}}
	r := httptest.NewRequest(
		http.MethodPost,
		"/admin/login",
		strings.NewReader(body.Encode()),
	)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.PostLoginHandler(store.AccessAdmin)(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestAdminLogin_SenhaVazia(t *testing.T) {
	h := newLoginHandler(nil, nil, nil)

	body := url.Values{"email": {"admin@test.com"}, "password": {""}}
	r := httptest.NewRequest(
		http.MethodPost,
		"/admin/login",
		strings.NewReader(body.Encode()),
	)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.PostLoginHandler(store.AccessAdmin)(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestAdminLogin_UsuarioNaoEncontrado(t *testing.T) {
	us := &mockUserStore{
		getUser: func(email string) (*store.User, error) {
			return nil, errors.New("not found")
		},
	}
	h := newLoginHandler(us, nil, nil)

	body := url.Values{
		"email":    {"inexistente@test.com"},
		"password": {"senha123"},
	}
	r := httptest.NewRequest(
		http.MethodPost,
		"/admin/login",
		strings.NewReader(body.Encode()),
	)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Host = "empresa.localhost"
	w := httptest.NewRecorder()

	h.PostLoginHandler(store.AccessAdmin)(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestAdminLogin_SenhaErrada(t *testing.T) {
	hash := hashedPassword(t, "correta")
	us := &mockUserStore{
		getUser: func(email string) (*store.User, error) {
			return &store.User{Email: email, Password: hash}, nil
		},
	}
	h := newLoginHandler(us, nil, nil)

	body := url.Values{"email": {"admin@test.com"}, "password": {"errada"}}
	r := httptest.NewRequest(
		http.MethodPost,
		"/admin/login",
		strings.NewReader(body.Encode()),
	)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Host = "empresa.localhost"
	w := httptest.NewRecorder()

	h.PostLoginHandler(store.AccessAdmin)(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestAdminLogin_Sucesso(t *testing.T) {
	senha := "senha123"
	var sessaoCriada store.Session
	us := &mockUserStore{
		getUser: func(email string) (*store.User, error) {
			return &store.User{
				ID:       1,
				Email:    email,
				Password: hashedPassword(t, senha),
				TenantID: 1,
			}, nil
		},
	}
	ts := &mockTenantStore{
		getTenantByID: func(id uint) (*store.Tenant, error) {
			return &store.Tenant{ID: 1, Slug: "empresa"}, nil
		},
	}
	ss := &mockSessionStore{
		createSession: func(_ *http.Request, _ http.ResponseWriter, s store.Session) error {
			sessaoCriada = s
			return nil
		},
	}
	h := newLoginHandler(us, ts, ss)

	body := url.Values{"email": {"admin@test.com"}, "password": {senha}}
	r := httptest.NewRequest(
		http.MethodPost,
		"/admin/login",
		strings.NewReader(body.Encode()),
	)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Host = "empresa.localhost"
	w := httptest.NewRecorder()

	h.PostLoginHandler(store.AccessAdmin)(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
	if w.Header().Get(HXRedirect) != "/admin/" {
		t.Errorf(
			"esperado redirect para /admin/, obteve %q",
			w.Header().Get(HXRedirect),
		)
	}
	if sessaoCriada.UserAccessType != store.AccessAdmin {
		t.Error("sessão criada com tipo de acesso incorreto")
	}
}

func TestAdminLogin_SlugDiferente(t *testing.T) {
	senha := "senha123"
	us := &mockUserStore{
		getUser: func(email string) (*store.User, error) {
			return &store.User{
				ID:       1,
				Email:    email,
				Password: hashedPassword(t, senha),
				TenantID: 1,
			}, nil
		},
	}
	ts := &mockTenantStore{
		getTenantByID: func(id uint) (*store.Tenant, error) {
			return &store.Tenant{
				ID:   1,
				Slug: "outro",
			}, nil // slug diferente do host
		},
	}
	h := newLoginHandler(us, ts, nil)

	body := url.Values{"email": {"admin@test.com"}, "password": {senha}}
	r := httptest.NewRequest(
		http.MethodPost,
		"/admin/login",
		strings.NewReader(body.Encode()),
	)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Host = "empresa.localhost" // slug "empresa" ≠ "outro"
	w := httptest.NewRecorder()

	h.PostLoginHandler(store.AccessAdmin)(w, r)

	// deve renderizar erro, não redirecionar
	if w.Header().Get(HXRedirect) == "/admin/" {
		t.Error("não deveria redirecionar quando slug não bate")
	}
}

func TestClientLogin_TelefoneVazio(t *testing.T) {
	h := newLoginHandler(nil, nil, nil)

	body := url.Values{"phone_number": {""}, "password": {"senha123"}}
	r := httptest.NewRequest(
		http.MethodPost,
		"/client/login",
		strings.NewReader(body.Encode()),
	)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.PostLoginHandler(store.AccessCustomer)(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestClientLogin_Sucesso(t *testing.T) {
	senha := "senha123"
	us := &mockUserStore{
		getUserByPhone: func(phone string) (*store.User, error) {
			return &store.User{
				ID:       1,
				Phone:    phone,
				Password: hashedPassword(t, senha),
				Contacts: []store.Contact{
					{ID: 5, TenantID: 1, PriceTableID: 2},
				},
			}, nil
		},
	}
	h := newLoginHandler(us, nil, nil)

	body := url.Values{"phone_number": {"11999999999"}, "password": {senha}}
	r := httptest.NewRequest(
		http.MethodPost,
		"/client/login",
		strings.NewReader(body.Encode()),
	)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Host = "empresa.localhost"
	w := httptest.NewRecorder()

	h.PostLoginHandler(store.AccessCustomer)(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
	if w.Header().Get(HXRedirect) != "/client/items" {
		t.Errorf(
			"esperado redirect para /client/items, obteve %q",
			w.Header().Get(HXRedirect),
		)
	}
}

func TestPostLogout_Admin(t *testing.T) {
	ss := &mockSessionStore{
		deleteSession: func(*http.Request, http.ResponseWriter) error { return nil },
	}
	h := newLoginHandler(nil, nil, ss)

	r := httptest.NewRequest(http.MethodPost, "/admin/logout", nil)
	w := httptest.NewRecorder()

	h.PostLogout(w, r)

	if w.Header().Get(HXRedirect) != "/admin/login" {
		t.Errorf(
			"esperado redirect para /admin/login, obteve %q",
			w.Header().Get(HXRedirect),
		)
	}
}

func TestPostLogout_Client(t *testing.T) {
	ss := &mockSessionStore{
		deleteSession: func(*http.Request, http.ResponseWriter) error { return nil },
	}
	h := newLoginHandler(nil, nil, ss)

	r := httptest.NewRequest(http.MethodPost, "/client/logout", nil)
	w := httptest.NewRecorder()

	h.PostLogout(w, r)

	if w.Header().Get(HXRedirect) != "/client/login" {
		t.Errorf(
			"esperado redirect para /client/login, obteve %q",
			w.Header().Get(HXRedirect),
		)
	}
}
