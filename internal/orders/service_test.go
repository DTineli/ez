package orders

import (
	"errors"
	"testing"
)

// mockRepo implements Repository for tests — only GetByID and Salvar matter here.
type mockRepo struct {
	pedido   *OrderDetail
	getErr   error
	salvarErr error
	saved    *OrderDetail
}

func (m *mockRepo) GetByID(id, tenantID uint) (*OrderDetail, error) {
	return m.pedido, m.getErr
}
func (m *mockRepo) Salvar(o *OrderDetail) error {
	m.saved = o
	return m.salvarErr
}
func (m *mockRepo) ConfirmFromCart(cartID, tenantID, contactID, priceTableID uint) (*Order, error) {
	return nil, nil
}
func (m *mockRepo) ListByTenant(tenantID uint) ([]AdminOrderListItem, error) { return nil, nil }
func (m *mockRepo) ListByTenantPaged(tenantID uint, filters OrderFilters) ([]AdminOrderListItem, int64, error) {
	return nil, 0, nil
}
func (m *mockRepo) ListByContact(tenantID, contactID uint) ([]ClientOrderListItem, error) {
	return nil, nil
}
func (m *mockRepo) Create(tenantID, contactID uint, items []NewOrderItem) (*Order, error) {
	return nil, nil
}

func pedidoWith(status, paymentStatus Status) *OrderDetail {
	return &OrderDetail{ID: 1, Status: status, PaymentStatus: paymentStatus}
}

// --- AtualizarStatus ---

func TestAtualizarStatus_TransicaoValida(t *testing.T) {
	repo := &mockRepo{pedido: pedidoWith(Pendente, PagamentoPendente)}
	svc := NewService(repo)

	err := svc.AtualizarStatus(1, 1, Aprovado, AtorSeller)
	if err != nil {
		t.Fatalf("esperava nil, got %v", err)
	}
	if repo.saved.Status != Aprovado {
		t.Errorf("status esperado %s, got %s", Aprovado, repo.saved.Status)
	}
}

func TestAtualizarStatus_TransicaoInvalida(t *testing.T) {
	repo := &mockRepo{pedido: pedidoWith(Pendente, PagamentoPendente)}
	svc := NewService(repo)

	err := svc.AtualizarStatus(1, 1, Entregue, AtorSeller)
	if err == nil {
		t.Fatal("esperava erro para transição inválida")
	}
}

func TestAtualizarStatus_GetByIDErro(t *testing.T) {
	repo := &mockRepo{getErr: errors.New("not found")}
	svc := NewService(repo)

	err := svc.AtualizarStatus(1, 1, Aprovado, AtorSeller)
	if err == nil {
		t.Fatal("esperava erro propagado de GetByID")
	}
}

func TestAtualizarStatus_SalvarErro(t *testing.T) {
	repo := &mockRepo{
		pedido:    pedidoWith(Pendente, PagamentoPendente),
		salvarErr: errors.New("db error"),
	}
	svc := NewService(repo)

	err := svc.AtualizarStatus(1, 1, Aprovado, AtorSeller)
	if err == nil {
		t.Fatal("esperava erro propagado de Salvar")
	}
}

func TestAtualizarStatus_EntregueSetaTimestamp(t *testing.T) {
	repo := &mockRepo{pedido: pedidoWith(AguardandoRetirada, PagamentoPendente)}
	svc := NewService(repo)

	err := svc.AtualizarStatus(1, 1, Entregue, AtorSeller)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.saved.EntregueEm == nil {
		t.Error("EntregueEm deveria ser preenchido")
	}
}

func TestAtualizarStatus_CanceladoSetaTimestamp(t *testing.T) {
	repo := &mockRepo{pedido: pedidoWith(Pendente, PagamentoPendente)}
	svc := NewService(repo)

	err := svc.AtualizarStatus(1, 1, Cancelado, AtorSeller)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.saved.CanceladoEm == nil {
		t.Error("CanceladoEm deveria ser preenchido")
	}
}

// --- tentarCompletar via AtualizarStatus ---

func TestAtualizarStatus_EntregueComPago_ViraCompleto(t *testing.T) {
	repo := &mockRepo{pedido: pedidoWith(AguardandoRetirada, Pago)}
	svc := NewService(repo)

	err := svc.AtualizarStatus(1, 1, Entregue, AtorSeller)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.saved.Status != Completo {
		t.Errorf("esperava %s, got %s", Completo, repo.saved.Status)
	}
}

func TestAtualizarStatus_EntregueComPagamentoPendente_NaoCompleta(t *testing.T) {
	repo := &mockRepo{pedido: pedidoWith(AguardandoRetirada, PagamentoPendente)}
	svc := NewService(repo)

	_ = svc.AtualizarStatus(1, 1, Entregue, AtorSeller)
	if repo.saved.Status == Completo {
		t.Error("não deveria virar Completo sem pagamento")
	}
}

// --- MarcarPago ---

func TestMarcarPago_SetaPaymentStatusEData(t *testing.T) {
	repo := &mockRepo{pedido: pedidoWith(Aprovado, PagamentoPendente)}
	svc := NewService(repo)

	err := svc.MarcarPago(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.saved.PaymentStatus != Pago {
		t.Errorf("esperava PaymentStatus=%s, got %s", Pago, repo.saved.PaymentStatus)
	}
	if repo.saved.PaymentDate == nil {
		t.Error("PaymentDate deveria ser preenchido")
	}
}

func TestMarcarPago_EntregueComPago_ViraCompleto(t *testing.T) {
	repo := &mockRepo{pedido: pedidoWith(Entregue, PagamentoPendente)}
	svc := NewService(repo)

	err := svc.MarcarPago(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.saved.Status != Completo {
		t.Errorf("esperava %s, got %s", Completo, repo.saved.Status)
	}
}

func TestMarcarPago_NaoEntregue_NaoCompleta(t *testing.T) {
	repo := &mockRepo{pedido: pedidoWith(Aprovado, PagamentoPendente)}
	svc := NewService(repo)

	_ = svc.MarcarPago(1, 1)
	if repo.saved.Status == Completo {
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
		pedido:    pedidoWith(Aprovado, PagamentoPendente),
		salvarErr: errors.New("db error"),
	}
	svc := NewService(repo)

	err := svc.MarcarPago(1, 1)
	if err == nil {
		t.Fatal("esperava erro propagado de Salvar")
	}
}
