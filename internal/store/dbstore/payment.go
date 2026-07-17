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

func (p *PaymentMethod) FindAllByPriceTable(
	tableID, tenantID uint,
) ([]store.PaymentMethod, error) {
	var pms []store.PaymentMethod
	err := p.db.
		Joins("JOIN price_table_payment_methods ptpm ON ptpm.payment_method_id = payment_methods.id").
		Where("ptpm.price_table_id = ? AND payment_methods.tenant_id = ?", tableID, tenantID).
		Find(&pms).Error
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
			"name": pm.Name,
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

type PaymentTerm struct {
	db *gorm.DB
}

func NewPaymentTermStore(db *gorm.DB) *PaymentTerm {
	return &PaymentTerm{
		db: db,
	}
}

func (p *PaymentTerm) CreatePaymentTerm(pt *store.PaymentTerm) error {
	return p.db.Create(pt).Error
}

func (p *PaymentTerm) GetPaymentTerm(id, tenantID uint) (*store.PaymentTerm, error) {
	var pt store.PaymentTerm
	err := p.db.Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&pt).
		Error
	if err != nil {
		return nil, err
	}
	return &pt, nil
}

func (p *PaymentTerm) FindAllPaymentTermsByTenant(
	tenantID uint,
) ([]store.PaymentTerm, error) {
	var pts []store.PaymentTerm
	err := p.db.Where("tenant_id = ?", tenantID).Find(&pts).Error
	if err != nil {
		return nil, err
	}
	return pts, nil
}

func (p *PaymentTerm) FindAllByPaymentMethod(
	methodID, tenantID uint,
) ([]store.PaymentTerm, error) {
	var pts []store.PaymentTerm
	err := p.db.
		Where("payment_method_id = ? AND tenant_id = ?", methodID, tenantID).
		Find(&pts).Error
	if err != nil {
		return nil, err
	}
	return pts, nil
}

func (p *PaymentTerm) UpdatePaymentTerm(pt *store.PaymentTerm) error {
	result := p.db.
		Model(&store.PaymentTerm{}).
		Where("id = ? AND tenant_id = ?", pt.ID, pt.TenantID).
		Updates(map[string]any{
			"name":              pt.Name,
			"due_days":          pt.DueDays,
			"percentage":        pt.Percentage,
			"payment_method_id": pt.PaymentMethodID,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("payment term not found")
	}

	return nil
}

func (p *PaymentTerm) DeletePaymentTerm(id, tenantID uint) error {
	result := p.db.
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&store.PaymentTerm{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("payment term not found")
	}

	return nil
}
