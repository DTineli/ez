package services

import "github.com/DTineli/ez/internal/store"

type PaymentMethodService interface {
	Create(tenantID uint, name string) (*store.PaymentMethod, error)
	GetOne(id, tenantID uint) (*store.PaymentMethod, error)
	FindAll(tenantID uint) ([]store.PaymentMethod, error)

	FindAllByPriceTable(tableID, tenantID uint) ([]store.PaymentMethod, error)
	Update(pm *store.PaymentMethod) error
	Delete(id, tenantID uint) error
}

type paymentMethodService struct {
	store store.PaymentMethodStore
}

func NewPaymentMethodService(s store.PaymentMethodStore) PaymentMethodService {
	return &paymentMethodService{store: s}
}

func (p *paymentMethodService) Create(
	tenantID uint,
	name string,
) (*store.PaymentMethod, error) {
	pm := &store.PaymentMethod{
		Name:     name,
		TenantID: tenantID,
	}
	if err := p.store.CreatePaymentMethod(pm); err != nil {
		return nil, err
	}
	return pm, nil
}

func (p *paymentMethodService) GetOne(id, tenantID uint) (*store.PaymentMethod, error) {
	return p.store.GetPaymentMethod(id, tenantID)
}

func (p *paymentMethodService) FindAll(tenantID uint) ([]store.PaymentMethod, error) {
	return p.store.FindAllPaymentMethodsByTenant(tenantID)
}

func (p *paymentMethodService) Update(pm *store.PaymentMethod) error {
	return p.store.UpdatePaymentMethod(pm)
}

func (p *paymentMethodService) Delete(id, tenantID uint) error {
	return p.store.DeletePaymentMethod(id, tenantID)
}

func (p *paymentMethodService) FindAllByPriceTable(
	tableID, tenantID uint,
) ([]store.PaymentMethod, error) {
	return p.store.FindAllByPriceTable(tableID, tenantID)
}
