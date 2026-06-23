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

func (p *PriceTableDB) CreateProductPrice(pPrice *store.ProductPrice) error {
	return p.db.Create(pPrice).Error
}

func (p *PriceTableDB) FindProductPrices(
	productID uint,
) ([]store.ProductPrice, error) {
	var prices []store.ProductPrice

	err := p.db.
		Joins("JOIN variants ON variants.id = product_prices.variant_id").
		Where("variants.product_id = ?", productID).
		Find(&prices).Error
	if err != nil {
		return nil, err
	}

	return prices, nil
}

func (p PriceTableDB) FindAllActiveByTenantAndClient(
	tenantID, clientID uint,
) ([]store.PriceTable, error) {
	var priceTables []store.PriceTable

	err := p.db.
		Joins("JOIN contact_price_tables cpt ON cpt.price_table_id = price_tables.id").
		Where("price_tables.status = true AND price_tables.tenant_id = ? AND cpt.contact_id = ?", tenantID, clientID).
		Find(&priceTables).
		Error
	if err != nil {
		return nil, err
	}

	return priceTables, nil
}

func (p PriceTableDB) FindAllActiveByTenant(
	id uint,
) ([]store.PriceTable, error) {
	var priceTables []store.PriceTable

	err := p.db.Where("status is true AND tenant_id = ?", id).
		Find(&priceTables).
		Error
	if err != nil {
		return nil, err
	}

	return priceTables, nil
}

func (p PriceTableDB) FindAllByTenant(id uint) ([]store.PriceTable, error) {
	var priceTables []store.PriceTable

	err := p.db.Where("tenant_id = ?", id).Find(&priceTables).Error
	if err != nil {
		return nil, err
	}

	return priceTables, nil
}

func (p PriceTableDB) GetOneWithPrices(id, tenantID uint) (*store.PriceTable, error) {
	var table store.PriceTable
	err := p.db.
		Preload("Prices.Variant.Product").
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&table).Error
	if err != nil {
		return nil, err
	}
	return &table, nil
}

func (p PriceTableDB) GetOne(
	id uint,
	tenantID uint,
) (*store.PriceTable, error) {
	var table store.PriceTable
	err := p.db.Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&table).
		Error
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

func (p *PriceTableDB) GetOneProductPrice(id uint) (*store.ProductPrice, error) {
	var price store.ProductPrice
	err := p.db.First(&price, id).Error
	if err != nil {
		return nil, err
	}
	return &price, nil
}

func (p *PriceTableDB) UpdateProductPrice(id uint, price float64) error {
	return p.db.Model(&store.ProductPrice{}).
		Where("id = ?", id).
		Update("price", price).Error
}

func (p *PriceTableDB) DeleteProductPrice(PriceID uint) error {
	return p.db.Delete(&store.ProductPrice{}, PriceID).Error
}
