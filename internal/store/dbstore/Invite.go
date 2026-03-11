package dbstore

import (
	"github.com/DTineli/ez/internal/store"
	"gorm.io/gorm"
)

type InviteStore struct {
	db *gorm.DB
}

func NewInvireStore(db *gorm.DB) *InviteStore {
	return &InviteStore{
		db: db,
	}
}

func (i *InviteStore) Create(invite *store.Invite) error {
	return i.db.Create(invite).Error
}
