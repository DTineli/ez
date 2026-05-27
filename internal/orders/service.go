package orders

import (
	"fmt"
	"time"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) AtualizarStatus(
	id, tenantID uint,
	para Status,
	ator Ator,
) error {
	pedido, err := s.repo.GetByID(id, tenantID)
	if err != nil {
		return err
	}

	if !PodeTransicionar(pedido.Status, para, ator) {
		return fmt.Errorf("transição inválida: %s → %s", pedido.Status, para)
	}

	pedido.Status = para

	switch para {
	case Entregue:
		now := time.Now()
		pedido.EntregueEm = &now
	case Cancelado:
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
	pedido.PaymentStatus = Pago
	pedido.PaymentDate = &now

	s.tentarCompletar(pedido) // mesma verificação

	return s.repo.Salvar(pedido)
}

func (s *Service) BulkAtualizarStatus(ids []uint, tenantID uint, status Status, ator Ator) {
	for _, id := range ids {
		_ = s.AtualizarStatus(id, tenantID, status, ator)
	}
}

func (s *Service) tentarCompletar(p *OrderDetail) {
	if p.Status == Entregue && p.PaymentStatus == Pago {
		p.Status = Completo
	}
}
