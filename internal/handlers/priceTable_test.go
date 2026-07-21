package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DTineli/ez/internal/services"
	"github.com/DTineli/ez/internal/store"
)

func newPriceTableHandler() *PriceTableHandler {
	return NewPriceTableHandler(
		&mockPriceTableServiceExt{},
		&mockPaymentMethodService{},
	)
}

func TestGetTablePage_Sucesso(t *testing.T) {
	h := newPriceTableHandler()

	r := httptest.NewRequest(http.MethodGet, "/admin/tabelas", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.GetTablePage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
}

func TestGetTablePage_ErroStore(t *testing.T) {
	pts := &mockPriceTableServiceExt{
		findAll: func(id uint) ([]store.PriceTable, error) {
			return nil, errors.New("db error")
		},
	}
	h := NewPriceTableHandler(pts, &mockPaymentMethodService{})

	r := httptest.NewRequest(http.MethodGet, "/admin/tabelas", nil)
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.GetTablePage(w, r)

	if !strings.Contains(w.Header().Get("HX-Trigger"), "error") {
		t.Error("esperado toast de erro")
	}
}

func TestCreatePriceTable_ValidacaoFalha_NomeFaltando(t *testing.T) {
	h := newPriceTableHandler()

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
	h := newPriceTableHandler()

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
	var capturedName string
	var capturedPct float64
	pts := &mockPriceTableServiceExt{
		create: func(tenantID uint, name string, pct float64) (*store.PriceTable, error) {
			capturedName = name
			capturedPct = pct
			return &store.PriceTable{Name: name, Percentage: pct, TenantID: tenantID}, nil
		},
	}
	h := NewPriceTableHandler(pts, &mockPaymentMethodService{})

	body := url.Values{"name": {"Tabela A"}, "percentage": {"10"}}
	r := httptest.NewRequest(http.MethodPost, "/admin/tabelas", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = htmxRequest(withSession(r, newSession(1)))
	w := httptest.NewRecorder()

	h.CreatePriceTable(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", w.Code)
	}
	if capturedName != "Tabela A" {
		t.Errorf("nome incorreto: %q", capturedName)
	}
	if capturedPct != 10 {
		t.Errorf("percentual incorreto: %v", capturedPct)
	}
	if !strings.Contains(w.Header().Get("HX-Trigger"), "success") {
		t.Error("esperado toast de sucesso")
	}
}

func TestCreatePriceTable_NomeDuplicado(t *testing.T) {
	pts := &mockPriceTableServiceExt{
		create: func(tenantID uint, name string, pct float64) (*store.PriceTable, error) {
			return nil, errors.New("UNIQUE constraint failed: price_tables.name")
		},
	}
	h := NewPriceTableHandler(pts, &mockPaymentMethodService{})

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
	h := newPriceTableHandler()

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
	pts := &mockPriceTableServiceExt{
		delete: func(id, tenantID uint) error {
			return services.ErrPriceTableHasContacts
		},
	}
	h := NewPriceTableHandler(pts, &mockPaymentMethodService{})

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
	pts := &mockPriceTableServiceExt{
		delete: func(id, tenantID uint) error {
			deleted = true
			return nil
		},
	}
	h := NewPriceTableHandler(pts, &mockPaymentMethodService{})

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
	if !strings.Contains(w.Header().Get("HX-Trigger"), "priceTableDeleted") {
		t.Errorf("esperado trigger priceTableDeleted, obteve %q", w.Header().Get("HX-Trigger"))
	}
}

// mockPriceTableServiceExt permite controle fino por teste
type mockPriceTableServiceExt struct {
	create        func(tenantID uint, name string, pct float64) (*store.PriceTable, error)
	delete        func(id, tenantID uint) error
	findAll       func(tenantID uint) ([]store.PriceTable, error)
	findAllActive func(tenantID uint) ([]store.PriceTable, error)
	getOne        func(id, tenantID uint) (*store.PriceTable, error)
}

func (s *mockPriceTableServiceExt) Create(tenantID uint, name string, pct float64) (*store.PriceTable, error) {
	if s.create != nil {
		return s.create(tenantID, name, pct)
	}
	return &store.PriceTable{Name: name, Percentage: pct}, nil
}
func (s *mockPriceTableServiceExt) Delete(id, tenantID uint) error {
	if s.delete != nil {
		return s.delete(id, tenantID)
	}
	return nil
}
func (s *mockPriceTableServiceExt) FindAll(tenantID uint) ([]store.PriceTable, error) {
	if s.findAll != nil {
		return s.findAll(tenantID)
	}
	return nil, nil
}
func (s *mockPriceTableServiceExt) FindAllActive(tenantID uint) ([]store.PriceTable, error) {
	if s.findAllActive != nil {
		return s.findAllActive(tenantID)
	}
	return nil, nil
}
func (s *mockPriceTableServiceExt) FindAllActiveByContact(tenantID, contactID uint) ([]store.PriceTable, error) {
	return nil, nil
}
func (s *mockPriceTableServiceExt) GetOne(id, tenantID uint) (*store.PriceTable, error) {
	if s.getOne != nil {
		return s.getOne(id, tenantID)
	}
	return &store.PriceTable{ID: id}, nil
}
func (s *mockPriceTableServiceExt) FindOne(id, tenantID uint) (*store.PriceTable, error) {
	return &store.PriceTable{ID: id}, nil
}
func (s *mockPriceTableServiceExt) Apply(costPrice float64, pt *store.PriceTable) float64 {
	return services.ApplyPriceTable(costPrice, pt)
}
func (s *mockPriceTableServiceExt) AddPrice(tableID, variationID uint, price float64) (uint, error) {
	return 0, nil
}
func (s *mockPriceTableServiceExt) GetVariantPrice(variantID, tableID uint) (float64, error) {
	return 0, nil
}
func (s *mockPriceTableServiceExt) GetProductPrice(id uint) (*store.ProductPrice, error) {
	return nil, nil
}
func (s *mockPriceTableServiceExt) UpdatePrice(id, tenantID uint, price float64) error { return nil }
func (s *mockPriceTableServiceExt) RemovePrice(priceID, tenantID uint) error           { return nil }
func (s *mockPriceTableServiceExt) SearchVariants(tenantID, priceTableID uint, q string) ([]store.Variant, error) {
	return nil, nil
}
func (s *mockPriceTableServiceExt) GetPriceTablesByProduct(productID, tenantID uint) ([]store.PriceTable, error) {
	return nil, nil
}
func (s *mockPriceTableServiceExt) FindAllWithProductPrices(productID, tenantID uint, allVariants []store.Variant) ([]store.PriceTableProductView, error) {
	return nil, nil
}
func (s *mockPriceTableServiceExt) FindPaymentMethods(tableID, tenantID uint) ([]store.PaymentMethod, error) {
	return nil, nil
}

func (s *mockPriceTableServiceExt) AddPaymentMethod(tableID, methodID, tenantID uint) error {
	return nil
}

func (s *mockPriceTableServiceExt) RemovePaymentMethod(tableID, methodID, tenantID uint) error {
	return nil
}

// mockPaymentMethodService implementa services.PaymentMethodService
type mockPaymentMethodService struct{}

func (s *mockPaymentMethodService) Create(tenantID uint, name string) (*store.PaymentMethod, error) {
	return &store.PaymentMethod{Name: name, TenantID: tenantID}, nil
}

func (s *mockPaymentMethodService) GetOne(id, tenantID uint) (*store.PaymentMethod, error) {
	return nil, nil
}

func (s *mockPaymentMethodService) FindAll(tenantID uint) ([]store.PaymentMethod, error) {
	return nil, nil
}

func (s *mockPaymentMethodService) FindAllByPriceTable(tableID, tenantID uint) ([]store.PaymentMethod, error) {
	return nil, nil
}

func (s *mockPaymentMethodService) Update(pm *store.PaymentMethod) error { return nil }
func (s *mockPaymentMethodService) Delete(id, tenantID uint) error       { return nil }

func (s *mockPaymentMethodService) CreateTerm(
	tenantID, methodID uint,
	dueDays int,
) (*store.PaymentTerm, error) {
	return &store.PaymentTerm{
		DueDays:         dueDays,
		PaymentMethodID: methodID,
		TenantID:        tenantID,
	}, nil
}

func (s *mockPaymentMethodService) FindTermsByMethod(methodID, tenantID uint) ([]store.PaymentTerm, error) {
	return nil, nil
}

func (s *mockPaymentMethodService) DeleteTerm(id, tenantID uint) error { return nil }
