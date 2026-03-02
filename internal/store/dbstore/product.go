package dbstore

import (
	"errors"

	"github.com/DTineli/ez/internal/store"
	"gorm.io/gorm"
)

type ProductStore struct {
	db *gorm.DB
}

func NewProductStore(db *gorm.DB) *ProductStore {
	return &ProductStore{
		db: db,
	}
}

func (p *ProductStore) CreateProduct(user *store.Product) error {
	return p.db.Create(user).Error
}

func (p *ProductStore) GetProduct(id uint) (*store.Product, error) {
	var product store.Product
	err := p.db.Where("id = ?", id).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (p *ProductStore) FindAllByUser(userID uint) ([]store.Product, error) {
	var products []store.Product

	err := p.db.Where("tenant_id = ?", userID).Find(&products).Error
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (p *ProductStore) UpdateById(product *store.Product) error {
	updates := map[string]any{
		"sku":              product.SKU,
		"name":             product.Name,
		"full_description": product.FullDescription,
		"status":           product.Status,
		"uom":              product.UOM,
		"ean":              product.EAN,
		"ncm":              product.NCM,
		"cost_price":       product.CostPrice,
		"width_cm":         product.WidthCm,
		"weight":           product.Weight,
		"height_cm":        product.HeightCm,
		"length_cm":        product.LengthCm,
		"minimum_stock":    product.MinimumStock,
		"current_stock":    product.CurrentStock,
	}

	result := p.db.
		Model(&store.Product{}).
		Where("id = ? AND tenant_id = ?", product.ID, product.TenantID).
		Updates(updates)

	if result.RowsAffected == 0 {
		return errors.New("product not found")
	}

	return result.Error
}

func (p ProductStore) FindAllByUserWithFilters(id uint, filters store.ProductFilters) ([]store.Product, error) {
	var products []store.Product
	if filters.SKU != "" {
		err := p.db.Where("tenant_id = ? AND sku = ?", id, filters.SKU).Find(&products).Error
		if err != nil {
			return nil, err
		}
		return products, nil
	}

	if filters.Name != "" {
		err := p.db.Where("tenant_id = ? AND name like %?%", id, filters.Name).Find(&products).Error
		if err != nil {
			return nil, err
		}
		return products, nil
	}

	err := p.db.Where("tenant_id = ?", id).Find(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (p *ProductStore) UpdateFields(
	id uint,
	tenantID uint,
	fields map[string]any,
) error {

	if len(fields) == 0 {
		return errors.New("no fields to update")
	}

	delete(fields, "id")
	delete(fields, "tenant_id")

	result := p.db.
		Model(&store.Product{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Updates(fields)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("product not found")
	}

	return nil
}
