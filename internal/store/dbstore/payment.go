package dbstore

import (
	"errors"

	"github.com/DTineli/ez/internal/store"
	"gorm.io/gorm"
)

type PaymentMethod struct {
	db *gorm.DB
}

func NewPaymentMethodStore(db *gorm.DB) *PaymentMethod {
	return &PaymentMethod{
		db: db,
	}
}

func (p *PaymentMethod) CreatePaymentMethod(pm *store.PaymentMethod) error {
	return p.db.Create(pm).Error
}

func (p *PaymentMethod) GetPaymentMethod(id, tenantID uint) (*store.PaymentMethod, error) {
	var pm store.PaymentMethod
	err := p.db.Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&pm).
		Error
	if err != nil {
		return nil, err
	}
	return &pm, nil
}

func (p *PaymentMethod) FindAllPaymentMethodsByTenant(
	tenantID uint,
) ([]store.PaymentMethod, error) {
	var pms []store.PaymentMethod
	err := p.db.Where("tenant_id = ?", tenantID).Find(&pms).Error
	if err != nil {
		return nil, err
	}
	return pms, nil
}

func (p *PaymentMethod) UpdatePaymentMethod(pm *store.PaymentMethod) error {
	result := p.db.
		Model(&store.PaymentMethod{}).
		Where("id = ? AND tenant_id = ?", pm.ID, pm.TenantID).
		Updates(map[string]any{
			"name":       pm.Name,
			"ativo":      pm.Ativo,
			"can_divide": pm.CanDivide,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("payment method not found")
	}

	return nil
}

func (p *PaymentMethod) DeletePaymentMethod(id, tenantID uint) error {
	result := p.db.
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&store.PaymentMethod{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("payment method not found")
	}

	return nil
}
