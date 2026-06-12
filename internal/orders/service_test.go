package orders

import (
	"errors"
	"testing"

	"github.com/DTineli/ez/internal/store"
)

type mockRepo struct {
	pedido    *store.OrderDetail
	getErr    error
	salvarErr error
	saved     *store.OrderDetail
}

func (m *mockRepo) GetByID(id, tenantID uint) (*store.OrderDetail, error) {
	return m.pedido, m.getErr
}
func (m *mockRepo) Salvar(o *store.OrderDetail) error {
	m.saved = o
	return m.salvarErr
}
func (m *mockRepo) ConfirmFromCart(cartID, tenantID, contactID, priceTableID uint) (*store.Order, error) {
	return nil, nil
}
func (m *mockRepo) ListByTenant(tenantID uint) ([]store.AdminOrderListItem, error) {
	return nil, nil
}
func (m *mockRepo) ListByTenantPaged(tenantID uint, filters store.OrderFilters) ([]store.AdminOrderListItem, int64, error) {
	return nil, 0, nil
}
func (m *mockRepo) ListByContact(tenantID, contactID uint) ([]store.ClientOrderListItem, error) {
	return nil, nil
}
func (m *mockRepo) Create(tenantID, contactID uint, items []store.NewOrderItem) (*store.Order, error) {
	return nil, nil
}

func pedidoWith(status, paymentStatus store.OrderStatus) *store.OrderDetail {
	return &store.OrderDetail{ID: 1, Status: status, PaymentStatus: paymentStatus}
}

// --- AtualizarStatus ---

func TestAtualizarStatus_TransicaoValida(t *testing.T) {
	repo := &mockRepo{pedido: pedidoWith(store.OrderPendente, store.OrderPagamentoPendente)}
	svc := NewService(repo)

	err := svc.AtualizarStatus(1, 1, store.OrderAprovado, store.OrderAtorSeller)
	if err != nil {
		t.Fatalf("esperava nil, got %v", err)
	}
	if repo.saved.Status != store.OrderAprovado {
		t.Errorf("status esperado %s, got %s", store.OrderAprovado, repo.saved.Status)
	}
}

func TestAtualizarStatus_TransicaoInvalida(t *testing.T) {
	repo := &mockRepo{pedido: pedidoWith(store.OrderPendente, store.OrderPagamentoPendente)}
	svc := NewService(repo)

	err := svc.AtualizarStatus(1, 1, store.OrderEntregue, store.OrderAtorSeller)
	if err == nil {
		t.Fatal("esperava erro para transição inválida")
	}
}

func TestAtualizarStatus_GetByIDErro(t *testing.T) {
	repo := &mockRepo{getErr: errors.New("not found")}
	svc := NewService(repo)

	err := svc.AtualizarStatus(1, 1, store.OrderAprovado, store.OrderAtorSeller)
	if err == nil {
		t.Fatal("esperava erro propagado de GetByID")
	}
}

func TestAtualizarStatus_SalvarErro(t *testing.T) {
	repo := &mockRepo{
		pedido:    pedidoWith(store.OrderPendente, store.OrderPagamentoPendente),
		salvarErr: errors.New("db error"),
	}
	svc := NewService(repo)

	err := svc.AtualizarStatus(1, 1, store.OrderAprovado, store.OrderAtorSeller)
	if err == nil {
		t.Fatal("esperava erro propagado de Salvar")
	}
}

func TestAtualizarStatus_EntregueSetaTimestamp(t *testing.T) {
	repo := &mockRepo{pedido: pedidoWith(store.OrderAguardandoRetirada, store.OrderPagamentoPendente)}
	svc := NewService(repo)

	err := svc.AtualizarStatus(1, 1, store.OrderEntregue, store.OrderAtorSeller)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.saved.EntregueEm == nil {
		t.Error("EntregueEm deveria ser preenchido")
	}
}

func TestAtualizarStatus_CanceladoSetaTimestamp(t *testing.T) {
	repo := &mockRepo{pedido: pedidoWith(store.OrderPendente, store.OrderPagamentoPendente)}
	svc := NewService(repo)

	err := svc.AtualizarStatus(1, 1, store.OrderCancelado, store.OrderAtorSeller)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.saved.CanceladoEm == nil {
		t.Error("CanceladoEm deveria ser preenchido")
	}
}

// --- tentarCompletar via AtualizarStatus ---

func TestAtualizarStatus_EntregueComPago_ViraCompleto(t *testing.T) {
	repo := &mockRepo{pedido: pedidoWith(store.OrderAguardandoRetirada, store.OrderPago)}
	svc := NewService(repo)

	err := svc.AtualizarStatus(1, 1, store.OrderEntregue, store.OrderAtorSeller)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.saved.Status != store.OrderCompleto {
		t.Errorf("esperava %s, got %s", store.OrderCompleto, repo.saved.Status)
	}
}

func TestAtualizarStatus_EntregueComPagamentoPendente_NaoCompleta(t *testing.T) {
	repo := &mockRepo{pedido: pedidoWith(store.OrderAguardandoRetirada, store.OrderPagamentoPendente)}
	svc := NewService(repo)

	_ = svc.AtualizarStatus(1, 1, store.OrderEntregue, store.OrderAtorSeller)
	if repo.saved.Status == store.OrderCompleto {
		t.Error("não deveria virar Completo sem pagamento")
	}
}

// --- MarcarPago ---

func TestMarcarPago_SetaPaymentStatusEData(t *testing.T) {
	repo := &mockRepo{pedido: pedidoWith(store.OrderAprovado, store.OrderPagamentoPendente)}
	svc := NewService(repo)

	err := svc.MarcarPago(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.saved.PaymentStatus != store.OrderPago {
		t.Errorf("esperava PaymentStatus=%s, got %s", store.OrderPago, repo.saved.PaymentStatus)
	}
	if repo.saved.PaymentDate == nil {
		t.Error("PaymentDate deveria ser preenchido")
	}
}

func TestMarcarPago_EntregueComPago_ViraCompleto(t *testing.T) {
	repo := &mockRepo{pedido: pedidoWith(store.OrderEntregue, store.OrderPagamentoPendente)}
	svc := NewService(repo)

	err := svc.MarcarPago(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.saved.Status != store.OrderCompleto {
		t.Errorf("esperava %s, got %s", store.OrderCompleto, repo.saved.Status)
	}
}

func TestMarcarPago_NaoEntregue_NaoCompleta(t *testing.T) {
	repo := &mockRepo{pedido: pedidoWith(store.OrderAprovado, store.OrderPagamentoPendente)}
	svc := NewService(repo)

	_ = svc.MarcarPago(1, 1)
	if repo.saved.Status == store.OrderCompleto {
		t.Error("não deveria virar Completo — pedido ainda não entregue")
	}
}

func TestMarcarPago_GetByIDErro(t *testing.T) {
	repo := &mockRepo{getErr: errors.New("not found")}
	svc := NewService(repo)

	err := svc.MarcarPago(1, 1)
	if err == nil {
		t.Fatal("esperava erro propagado de GetByID")
	}
}

func TestMarcarPago_SalvarErro(t *testing.T) {
	repo := &mockRepo{
		pedido:    pedidoWith(store.OrderAprovado, store.OrderPagamentoPendente),
		salvarErr: errors.New("db error"),
	}
	svc := NewService(repo)

	err := svc.MarcarPago(1, 1)
	if err == nil {
		t.Fatal("esperava erro propagado de Salvar")
	}
}
