package dbstore

import (
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
	return nil
}
