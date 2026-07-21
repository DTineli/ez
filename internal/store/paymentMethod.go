package store

import "gorm.io/gorm"

type PaymentMethod struct {
	gorm.Model
	Name string `gorm:"type:varchar(100);not null;uniqueIndex:idx_tenant_pm_name,priority:2,where:deleted_at IS NULL" json:"name"`

	TenantID uint          `gorm:"not null;uniqueIndex:idx_tenant_pm_name,priority:1" json:"tenant_id"`
	Terms    []PaymentTerm `gorm:"foreignKey:PaymentMethodID" json:"terms,omitempty"`
}

type PaymentTerm struct {
	gorm.Model
	DueDays int `json:"due_days"`

	PaymentMethodID uint          `gorm:"not null" json:"payment_method_id"`
	PaymentMethod   PaymentMethod `gorm:"foreignKey:PaymentMethodID" json:"-"`
	TenantID        uint          `gorm:"not null" json:"tenant_id"`
}

type PaymentMethodStore interface {
	CreatePaymentMethod(pm *PaymentMethod) error
	GetPaymentMethod(id, tenantID uint) (*PaymentMethod, error)
	FindAllPaymentMethodsByTenant(tenantID uint) ([]PaymentMethod, error)
	FindAllByPriceTable(tableID, tenantID uint) ([]PaymentMethod, error)
	UpdatePaymentMethod(pm *PaymentMethod) error
	DeletePaymentMethod(id, tenantID uint) error
}

type PaymentTermStore interface {
	CreatePaymentTerm(pt *PaymentTerm) error
	GetPaymentTerm(id, tenantID uint) (*PaymentTerm, error)
	FindAllPaymentTermsByTenant(tenantID uint) ([]PaymentTerm, error)
	FindAllByPaymentMethod(methodID, tenantID uint) ([]PaymentTerm, error)
	UpdatePaymentTerm(pt *PaymentTerm) error
	DeletePaymentTerm(id, tenantID uint) error
}
