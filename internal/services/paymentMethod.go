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
	methodStore store.PaymentMethodStore
	termStore   store.PaymentTermStore
}

func NewPaymentMethodService(mStore store.PaymentMethodStore,
	tStore store.PaymentTermStore) PaymentMethodService {
	return &paymentMethodService{methodStore: mStore, termStore: tStore}
}

func (p *paymentMethodService) Create(
	tenantID uint,
	name string,
) (*store.PaymentMethod, error) {
	pm := &store.PaymentMethod{
		Name:     name,
		TenantID: tenantID,
	}
	if err := p.methodStore.CreatePaymentMethod(pm); err != nil {
		return nil, err
	}
	return pm, nil
}

func (p *paymentMethodService) GetOne(id, tenantID uint) (*store.PaymentMethod, error) {
	return p.methodStore.GetPaymentMethod(id, tenantID)
}

func (p *paymentMethodService) FindAll(tenantID uint) ([]store.PaymentMethod, error) {
	return p.methodStore.FindAllPaymentMethodsByTenant(tenantID)
}

func (p *paymentMethodService) Update(pm *store.PaymentMethod) error {
	return p.methodStore.UpdatePaymentMethod(pm)
}

func (p *paymentMethodService) Delete(id, tenantID uint) error {
	return p.methodStore.DeletePaymentMethod(id, tenantID)
}

func (p *paymentMethodService) FindAllByPriceTable(
	tableID, tenantID uint,
) ([]store.PaymentMethod, error) {
	return p.methodStore.FindAllByPriceTable(tableID, tenantID)
}
