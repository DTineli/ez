package dbstore

import (
	"errors"
	"strings"

	"github.com/DTineli/ez/internal/store"
)

func (p *ProductStore) CreateAttribute(attr *store.Attribute) error {
	return p.db.Create(attr).Error
}

func (p *ProductStore) GetAttribute(
	id uint,
	tenantID uint,
) (*store.Attribute, error) {
	var attr store.Attribute
	err := p.db.Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&attr).
		Error
	if err != nil {
		return nil, err
	}
	return &attr, nil
}

func (p *ProductStore) FindAttributesByTenant(
	tenantID uint,
) ([]store.Attribute, error) {
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

func (p *ProductStore) AttributeInUse(id uint, tenantID uint) (bool, error) {
	var count int64
	err := p.db.Model(&store.VariantAttribute{}).
		Joins("JOIN attribute_values ON variant_attributes.attribute_value_id = attribute_values.id").
		Joins("JOIN attributes ON attribute_values.attribute_id = attributes.id").
		Where("attributes.id = ? AND attributes.tenant_id = ?", id, tenantID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (p *ProductStore) CreateAttributeValue(val *store.AttributeValue) error {
	return p.db.Create(val).Error
}

func (p *ProductStore) FindOrCreateAttribute(
	name string,
	tenantID uint,
) (*store.Attribute, error) {
	var attr store.Attribute
	name = strings.ToLower(strings.TrimSpace(name))
	result := p.db.Where(store.Attribute{Name: name, TenantID: tenantID}).
		FirstOrCreate(&attr)
	return &attr, result.Error
}

func (p *ProductStore) FindOrCreateAttributeValue(
	value string,
	attrID uint,
) (*store.AttributeValue, error) {
	var av store.AttributeValue
	result := p.db.Where(store.AttributeValue{Value: value, AttributeID: attrID}).
		FirstOrCreate(&av)
	return &av, result.Error
}

func (p *ProductStore) DeleteAttributeValue(id uint, tenantID uint) error {
	result := p.db.
		Where(
			"id = ? AND attribute_id IN (SELECT id FROM attributes WHERE tenant_id = ?)",
			id,
			tenantID,
		).
		Delete(&store.AttributeValue{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("attribute value not found")
	}

	return nil
}
