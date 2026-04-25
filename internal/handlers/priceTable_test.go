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

func TestGetTablePage_Sucesso(t *testing.T) {
	h := newHandler()

	r := httptest.NewRequest(http.MethodGet, "/admin/tabelas", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.GetTablePage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestGetTablePage_ErroStore(t *testing.T) {
	pts := &mockPriceTableStoreExt{
		findAllByTenant: func(id uint) ([]store.PriceTable, error) {
			return nil, errors.New("db error")
		},
	}
	h := NewProductHandler(&mockProductStore{}, pts)

	r := httptest.NewRequest(http.MethodGet, "/admin/tabelas", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.GetTablePage(w, r)

	// ShowToast escreve WriteHeader(200) antes de http.Error, então status é 200
	// mas o toast de erro deve estar presente
	if !strings.Contains(w.Header().Get("HX-Trigger"), "error") {
		t.Error("esperado toast de erro")
	}
}

func TestCreatePriceTable_ValidacaoFalha_NomeFaltando(t *testing.T) {
	h := newHandler()

	body := url.Values{"percentage": {"10"}} // sem name
	r := httptest.NewRequest(http.MethodPost, "/admin/tabelas", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.CreatePriceTable(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
	if !strings.Contains(w.Header().Get("HX-Trigger"), "error") {
		t.Error("esperado toast de erro na validação")
	}
}

func TestCreatePriceTable_ValidacaoFalha_PercentualInvalido(t *testing.T) {
	h := newHandler()

	body := url.Values{"name": {"Tabela A"}, "percentage": {"abc"}}
	r := httptest.NewRequest(http.MethodPost, "/admin/tabelas", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.CreatePriceTable(w, r)

	if !strings.Contains(w.Header().Get("HX-Trigger"), "error") {
		t.Error("esperado toast de erro na validação")
	}
}

func TestCreatePriceTable_Sucesso(t *testing.T) {
	var criada *store.PriceTable
	pts := &mockPriceTableStoreExt{
		createPriceTable: func(p *store.PriceTable) error {
			criada = p
			return nil
		},
	}
	h := NewProductHandler(&mockProductStore{}, pts)

	body := url.Values{"name": {"Tabela A"}, "percentage": {"10"}}
	r := httptest.NewRequest(http.MethodPost, "/admin/tabelas", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.CreatePriceTable(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
	if criada == nil {
		t.Fatal("CreatePriceTable não foi chamado")
	}
	if criada.Name != "Tabela A" {
		t.Errorf("nome incorreto: %q", criada.Name)
	}
	if criada.Percentage != 10 {
		t.Errorf("percentual incorreto: %v", criada.Percentage)
	}
	if !strings.Contains(w.Header().Get("HX-Trigger"), "success") {
		t.Error("esperado toast de sucesso")
	}
}

func TestCreatePriceTable_NomeDuplicado(t *testing.T) {
	pts := &mockPriceTableStoreExt{
		createPriceTable: func(p *store.PriceTable) error {
			return errors.New("UNIQUE constraint failed: price_tables.name")
		},
	}
	h := NewProductHandler(&mockProductStore{}, pts)

	body := url.Values{"name": {"Tabela Existente"}, "percentage": {"5"}}
	r := httptest.NewRequest(http.MethodPost, "/admin/tabelas", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.CreatePriceTable(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
	if !strings.Contains(w.Header().Get("HX-Trigger"), "error") {
		t.Error("esperado toast de erro para nome duplicado")
	}
}

func TestDeletePriceTable_IDInvalido(t *testing.T) {
	h := newHandler()

	r := httptest.NewRequest(http.MethodDelete, "/admin/tabelas/abc", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParam(r, "id", "abc")
	w := httptest.NewRecorder()

	h.DeletePriceTable(w, r)

	if !strings.Contains(w.Header().Get("HX-Trigger"), "error") {
		t.Error("esperado toast de erro para ID inválido")
	}
}

func TestDeletePriceTable_PossuiClientes(t *testing.T) {
	pts := &mockPriceTableStoreExt{
		hasContacts: func(priceTableID, tenantID uint) (bool, error) {
			return true, nil
		},
	}
	h := NewProductHandler(&mockProductStore{}, pts)

	r := httptest.NewRequest(http.MethodDelete, "/admin/tabelas/1", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParam(r, "id", "1")
	w := httptest.NewRecorder()

	h.DeletePriceTable(w, r)

	if !strings.Contains(w.Header().Get("HX-Trigger"), "error") {
		t.Error("esperado toast de erro: tabela com clientes")
	}
}

func TestDeletePriceTable_Sucesso(t *testing.T) {
	deleted := false
	pts := &mockPriceTableStoreExt{
		hasContacts: func(priceTableID, tenantID uint) (bool, error) { return false, nil },
		delete: func(id, tenantID uint) error {
			deleted = true
			return nil
		},
	}
	h := NewProductHandler(&mockProductStore{}, pts)

	r := httptest.NewRequest(http.MethodDelete, "/admin/tabelas/1", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	r = withChiParam(r, "id", "1")
	w := httptest.NewRecorder()

	h.DeletePriceTable(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
	if !deleted {
		t.Error("Delete não foi chamado")
	}
	// handler sobrescreve HX-Trigger com priceTableDeleted após ShowToast
	if !strings.Contains(w.Header().Get("HX-Trigger"), "priceTableDeleted") {
		t.Errorf("esperado trigger priceTableDeleted, obteve %q", w.Header().Get("HX-Trigger"))
	}
}

// mockPriceTableStoreExt permite controle fino por teste sem sobrescrever o stub global
type mockPriceTableStoreExt struct {
	createPriceTable func(*store.PriceTable) error
	findAllByTenant  func(id uint) ([]store.PriceTable, error)
	getOne           func(id, tenantID uint) (*store.PriceTable, error)
	hasContacts      func(priceTableID, tenantID uint) (bool, error)
	delete           func(id, tenantID uint) error
}

func (s *mockPriceTableStoreExt) CreatePriceTable(p *store.PriceTable) error {
	if s.createPriceTable != nil {
		return s.createPriceTable(p)
	}
	return nil
}
func (s *mockPriceTableStoreExt) FindAllByTenant(id uint) ([]store.PriceTable, error) {
	if s.findAllByTenant != nil {
		return s.findAllByTenant(id)
	}
	return nil, nil
}
func (s *mockPriceTableStoreExt) GetOne(id, tenantID uint) (*store.PriceTable, error) {
	if s.getOne != nil {
		return s.getOne(id, tenantID)
	}
	return &store.PriceTable{ID: id}, nil
}
func (s *mockPriceTableStoreExt) HasContacts(priceTableID, tenantID uint) (bool, error) {
	if s.hasContacts != nil {
		return s.hasContacts(priceTableID, tenantID)
	}
	return false, nil
}
func (s *mockPriceTableStoreExt) Delete(id, tenantID uint) error {
	if s.delete != nil {
		return s.delete(id, tenantID)
	}
	return nil
}
