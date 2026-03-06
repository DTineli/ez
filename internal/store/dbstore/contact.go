package dbstore

import (
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

func (c *ContactStore) FindAll(tenantID uint, filters store.ContactFilters) (*store.FindResults[store.Contact], error) {
	var contacts []store.Contact
	query := c.db.Model(&store.Contact{}).Where("tenant_id = ?", tenantID)

	query = ApplyFilters(query, filters)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return nil, err
	}

	// Paginação + Ordenação
	query = query.Order("id DESC").Offset((filters.Page - 1) * filters.PerPage).Limit(filters.PerPage)

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
			query = query.Where(column+" LIKE ?", "%"+field.String()+"%")
		case "eq":
			query = query.Where(column+" = ?", field.Interface())
		}
	}

	return query
}
