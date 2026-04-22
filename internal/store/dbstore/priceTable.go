package dbstore

import (
	"github.com/DTineli/ez/internal/store"
	"gorm.io/gorm"
)

type PriceTableDB struct {
	db *gorm.DB
}

func NewPriceTableDB(db *gorm.DB) *PriceTableDB {
	return &PriceTableDB{
		db: db,
	}
}

func (p PriceTableDB) CreatePriceTable(table *store.PriceTable) error {
	return p.db.Create(table).Error
}

func (p PriceTableDB) FindAllByTenant(id uint) ([]store.PriceTable, error) {
	var priceTables []store.PriceTable

	err := p.db.Where("tenant_id = ?", id).Find(&priceTables).Error
	if err != nil {
		return nil, err
	}

	return priceTables, nil
}

func (p PriceTableDB) GetOne(id uint, tenantID uint) (*store.PriceTable, error) {
	var table store.PriceTable
	err := p.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&table).Error
	if err != nil {
		return nil, err
	}
	return &table, nil
}

func (p PriceTableDB) HasContacts(priceTableID, tenantID uint) (bool, error) {
	var count int64
	err := p.db.Model(&store.Contact{}).
		Where("price_table_id = ? AND tenant_id = ?", priceTableID, tenantID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (p PriceTableDB) Delete(id, tenantID uint) error {
	return p.db.
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&store.PriceTable{}).Error
}
