package orders

import (
	"fmt"
	"time"

	"github.com/DTineli/ez/internal/store"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) FetchOrderInfo(
	ids []uint,
	tennantID uint,
) ([]store.OrderDetail, error) {

	return []store.OrderDetail{}, nil
}

func (s *Service) AtualizarStatus(
	id, tenantID uint,
	para store.OrderStatus,
	ator store.OrderAtor,
) error {
	pedido, err := s.repo.GetByID(id, tenantID)
	if err != nil {
		return err
	}

	if !store.PodeTransicionarOrder(pedido.Status, para, ator) {
		return fmt.Errorf("transição inválida: %s → %s", pedido.Status, para)
	}

	pedido.Status = para

	switch para {
	case store.OrderEntregue:
		now := time.Now()
		pedido.EntregueEm = &now
	case store.OrderCancelado:
		now := time.Now()
		pedido.CanceladoEm = &now
	}

	s.tentarCompletar(pedido)

	return s.repo.Salvar(pedido)
}

func (s *Service) MarcarPago(id, tenantID uint) error {
	pedido, err := s.repo.GetByID(id, tenantID)
	if err != nil {
		return err
	}

	now := time.Now()
	pedido.PaymentStatus = store.OrderPago
	pedido.PaymentDate = &now

	s.tentarCompletar(pedido)

	return s.repo.Salvar(pedido)
}

func (s *Service) BulkAtualizarStatus(
	ids []uint,
	tenantID uint,
	status store.OrderStatus,
	ator store.OrderAtor,
) {
	for _, id := range ids {
		_ = s.AtualizarStatus(id, tenantID, status, ator)
	}
}

func (s *Service) tentarCompletar(p *store.OrderDetail) {
	if p.Status == store.OrderEntregue && p.PaymentStatus == store.OrderPago {
		p.Status = store.OrderCompleto
	}
}
