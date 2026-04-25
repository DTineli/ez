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

func (p *ProductStore) CreateProduct(product *store.Product) error {
	return p.db.Create(product).Error
}

func (p *ProductStore) GetProduct(id uint) (*store.Product, error) {
	var product store.Product
	err := p.db.Preload("Variants").Where("id = ?", id).First(&product).Error
	return &product, err
}

func (p *ProductStore) FindAllByUser(userID uint) ([]store.Product, error) {
	var products []store.Product

	err := p.db.Where("tenant_id = ?", userID).Find(&products).Error
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (p *ProductStore) FindAllByUserWithFilters(id uint, filters store.ProductFilters) (*store.FindResults[store.Product], error) {
	var products []store.Product
	query := p.db.Model(&store.Product{}).Preload("Variants").Where("tenant_id = ?", id)

	if filters.Search != "" {
		like := "%" + filters.Search + "%"
		query = query.Where("name LIKE ? OR sku LIKE ?", like, like)
	} else {
		if filters.SKU != "" {
			query = query.Where("sku = ?", filters.SKU)
		}
		if filters.Name != "" {
			query = query.Where("name LIKE ?", "%"+filters.Name+"%")
		}
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return nil, err
	}

	query = query.Order("id DESC").Offset((filters.Page - 1) * filters.PerPage).Limit(filters.PerPage)

	if err := query.Find(&products).Error; err != nil {
		return nil, err
	}

	return &store.FindResults[store.Product]{
		Count:   count,
		Results: products,
	}, nil
}

func (p *ProductStore) UpdateFields(id uint, tenantID uint, fields map[string]any) error {
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

// --- Variant ---

func (p *ProductStore) CreateVariant(variant *store.Variant) error {
	return p.db.Create(variant).Error
}

func (p *ProductStore) GetVariant(id uint, tenantID uint) (*store.Variant, error) {
	var variant store.Variant
	err := p.db.
		Preload("Attributes.AttributeValue.Attribute").
		Preload("Prices.PriceTable").
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&variant).Error
	if err != nil {
		return nil, err
	}
	return &variant, nil
}

func (p *ProductStore) FindVariantsByProduct(productID uint, tenantID uint) ([]store.Variant, error) {
	var variants []store.Variant
	err := p.db.
		Preload("Attributes.AttributeValue.Attribute").
		Where("product_id = ? AND tenant_id = ?", productID, tenantID).
		Find(&variants).Error
	if err != nil {
		return nil, err
	}
	return variants, nil
}

func (p *ProductStore) UpdateVariantFields(id uint, tenantID uint, fields map[string]any) error {
	if len(fields) == 0 {
		return errors.New("no fields to update")
	}

	delete(fields, "id")
	delete(fields, "tenant_id")
	delete(fields, "product_id")

	result := p.db.
		Model(&store.Variant{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Updates(fields)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("variant not found")
	}

	return nil
}

func (p *ProductStore) DeleteVariant(id uint, tenantID uint) error {
	result := p.db.
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&store.Variant{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("variant not found")
	}

	return nil
}

// SetVariantAttributes substitui todos os atributos do variant atomicamente.
func (p *ProductStore) SetVariantAttributes(variantID uint, attributeValueIDs []uint) error {
	return p.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("variant_id = ?", variantID).Delete(&store.VariantAttribute{}).Error; err != nil {
			return err
		}

		for _, avID := range attributeValueIDs {
			va := store.VariantAttribute{
				VariantID:        variantID,
				AttributeValueID: avID,
			}
			if err := tx.Create(&va).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// --- Attribute ---

func (p *ProductStore) CreateAttribute(attr *store.Attribute) error {
	return p.db.Create(attr).Error
}

func (p *ProductStore) GetAttribute(id uint, tenantID uint) (*store.Attribute, error) {
	var attr store.Attribute
	err := p.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&attr).Error
	if err != nil {
		return nil, err
	}
	return &attr, nil
}

func (p *ProductStore) FindAttributesByTenant(tenantID uint) ([]store.Attribute, error) {
	var attrs []store.Attribute
	err := p.db.
		Preload("Values").
		Where("tenant_id = ?", tenantID).
		Find(&attrs).Error
	if err != nil {
		return nil, err
	}
	return attrs, nil
}

func (p *ProductStore) DeleteAttribute(id uint, tenantID uint) error {
	result := p.db.
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&store.Attribute{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("attribute not found")
	}

	return nil
}

func (p *ProductStore) CreateAttributeValue(val *store.AttributeValue) error {
	return p.db.Create(val).Error
}

func (p *ProductStore) DeleteAttributeValue(id uint, tenantID uint) error {
	// AttributeValue não tem tenant_id direto; valida via join com Attribute
	result := p.db.
		Where("id = ? AND attribute_id IN (SELECT id FROM attributes WHERE tenant_id = ?)", id, tenantID).
		Delete(&store.AttributeValue{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("attribute value not found")
	}

	return nil
}
