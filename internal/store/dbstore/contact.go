package dbstore

import (
	"errors"
	"reflect"
	"strings"

	"github.com/DTineli/ez/internal/store"
	"gorm.io/gorm"
)

type ContactStore struct {
	db *gorm.DB
}

func NewContactStore(db *gorm.DB) *ContactStore {
	return &ContactStore{
		db: db,
	}
}

func (c *ContactStore) CreateContact(contact *store.Contact) error {
	return c.db.Create(contact).Error
}

func (c *ContactStore) GetOne(id uint) (*store.Contact, error) {
	var contact = &store.Contact{}
	err := c.db.Preload("PriceTables").Where("id = ?", id).Find(&contact).Error
	return contact, err
}

func (c *ContactStore) FindContactPriceTables(
	contactID,
	tenantID uint,
) ([]store.PriceTable, error) {
	var pricetables []store.PriceTable

	c.db.Joins("JOIN contact_price_tables cpt ON price_tables.id = cpt.price_table_id").
		Where("cpt.contact_id = ? AND price_tables.status is true", contactID).
		Where("price_tables.tenant_id = ?", tenantID).
		Find(&pricetables)

	return pricetables, nil
}

func (c *ContactStore) UpdateById(
	id uint,
	tenantID uint,
	fields map[string]any,
) error {
	if len(fields) == 0 {
		return errors.New("no fields to update")
	}

	delete(fields, "id")
	delete(fields, "tenant_id")

	// Extract price_table_ids before scalar update
	var priceTableIDs []uint
	if raw, ok := fields["price_table_ids"]; ok {
		delete(fields, "price_table_ids")
		if ids, ok := raw.([]uint); ok {
			priceTableIDs = ids
		}
	}

	if len(fields) > 0 {
		result := c.db.
			Model(&store.Contact{}).
			Where("id = ? AND tenant_id = ?", id, tenantID).
			Updates(fields)

		if result.RowsAffected == 0 {
			return errors.New("contact not found")
		}
		if result.Error != nil {
			return result.Error
		}
	}

	if priceTableIDs != nil {
		var priceTables []store.PriceTable
		if len(priceTableIDs) > 0 {
			if err := c.db.Where("id IN ? AND tenant_id = ?", priceTableIDs, tenantID).Find(&priceTables).Error; err != nil {
				return err
			}
		}
		contact := &store.Contact{}
		contact.ID = id
		if err := c.db.Model(contact).Association("PriceTables").Replace(priceTables); err != nil {
			return err
		}
	}

	return nil
}

func (c *ContactStore) FindAll(
	tenantID uint,
	filters store.ContactFilters,
) (*store.FindResults[store.Contact], error) {
	var contacts []store.Contact
	query := c.db.Model(&store.Contact{}).Where("tenant_id = ?", tenantID)

	query = ApplyFilters(query, filters)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return nil, err
	}

	// Paginação + Ordenação
	query = query.Order("id DESC").
		Offset((filters.Page - 1) * filters.PerPage).
		Limit(filters.PerPage)

	if err := query.Find(&contacts).Error; err != nil {
		return nil, err
	}

	return &store.FindResults[store.Contact]{
		Count:   0,
		Results: contacts,
	}, nil
}

func ApplyFilters(query *gorm.DB, filters any) *gorm.DB {
	v := reflect.ValueOf(filters)
	t := reflect.TypeOf(filters)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		tag := t.Field(i).Tag.Get("filter")

		if tag == "" || field.IsZero() {
			continue
		}

		parts := strings.Split(tag, ",")
		column := parts[0]
		op := parts[1]

		switch op {
		case "like":
			query = query.Where("LOWER("+column+") LIKE LOWER(?)", "%"+field.String()+"%")
		case "eq":
			query = query.Where(column+" = ?", field.Interface())
		}
	}

	return query
}
