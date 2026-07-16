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
	Name string `gorm:"type:varchar(100);not null;uniqueIndex:idx_tenant_pt_name,priority:2,where:deleted_at IS NULL" json:"name"`

	Number     int     `json:"number"`
	DueDays    int     `json:"due_days"`
	Percentage float64 `json:"percentage"`

	PaymentMethodID uint          `gorm:"not null" json:"payment_method_id"`
	PaymentMethod   PaymentMethod `gorm:"foreignKey:PaymentMethodID" json:"-"`
	TenantID        uint          `gorm:"not null;uniqueIndex:idx_tenant_pt_name,priority:1" json:"tenant_id"`
}

type PaymentMethodStore interface {
	CreatePaymentMethod(pm *PaymentMethod) error
	GetPaymentMethod(id, tenantID uint) (*PaymentMethod, error)
	FindAllPaymentMethodsByTenant(tenantID uint) ([]PaymentMethod, error)
	FindAllByPriceTable(tableID, tenantID uint) ([]PaymentMethod, error)
	UpdatePaymentMethod(pm *PaymentMethod) error
	DeletePaymentMethod(id, tenantID uint) error
}
