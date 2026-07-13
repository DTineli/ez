package dbstore

import (
	"errors"

	"github.com/DTineli/ez/internal/store"
	"gorm.io/gorm"
)

func (p *ProductStore) CreateVariant(variant *store.Variant) error {
	return p.db.Create(variant).Error
}

func (p *ProductStore) GetVariant(
	id uint,
	tenantID uint,
) (*store.Variant, error) {
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

func (p *ProductStore) GetVariantForCart(variantID, productID, tenantID uint) (*store.Variant, error) {
	var v store.Variant
	err := p.db.
		Where("id = ? AND product_id = ? AND tenant_id = ?", variantID, productID, tenantID).
		First(&v).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (p *ProductStore) FindVariantsByProduct(
	productID uint,
	tenantID uint,
) ([]store.Variant, error) {
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

func (p *ProductStore) UpdateVariantFields(
	id uint,
	tenantID uint,
	fields map[string]any,
) error {
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

func (p *ProductStore) FindDefaultVariant(
	productID uint,
	tenantID uint,
) (*store.Variant, error) {
	var v store.Variant
	result := p.db.
		Where(
			"product_id = ? AND tenant_id = ? AND is_default = ?",
			productID,
			tenantID,
			true,
		).
		First(&v)
	if result.Error != nil {
		return nil, result.Error
	}
	return &v, nil
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

// CreateVariants cria em lote as Variants (e seus VariantAttribute) descritas
// em inputs, tudo em uma única transação — se qualquer combo falhar (ex: SKU
// duplicado), nada é persistido.
func (p *ProductStore) CreateVariants(
	productID uint,
	tenantID uint,
	inputs []store.VariantGenInput,
) ([]store.Variant, error) {
	created := make([]store.Variant, 0, len(inputs))

	err := p.db.Transaction(func(tx *gorm.DB) error {
		for _, in := range inputs {
			v := store.Variant{
				SKU:          in.SKU,
				CostPrice:    in.CostPrice,
				CurrentStock: in.CurrentStock,
				ProductID:    productID,
				TenantID:     tenantID,
			}
			if err := tx.Create(&v).Error; err != nil {
				return err
			}

			for _, avID := range in.AttributeValueIDs {
				va := store.VariantAttribute{
					VariantID:        v.ID,
					AttributeValueID: avID,
				}
				if err := tx.Create(&va).Error; err != nil {
					return err
				}
			}

			created = append(created, v)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (p *ProductStore) SetVariantAttributes(
	variantID uint,
	attributeValueIDs []uint,
) error {
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
