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
	return nil, nil
}
