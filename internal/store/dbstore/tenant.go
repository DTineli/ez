package dbstore

import (
	"github.com/DTineli/ez/internal/store"
	"gorm.io/gorm"
)

type TenantStore struct {
	db *gorm.DB
}

func NewTenantStore(db *gorm.DB) *TenantStore {
	return &TenantStore{
		db: db,
	}
}

func (t TenantStore) CreateTenant(tenant store.Tenant) (uint, error) {
	queryresut := t.db.Create(&tenant)

	return tenant.ID, queryresut.Error
}
