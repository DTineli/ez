package dbstore

import (
	"github.com/DTineli/ez/internal/store"
	"github.com/google/uuid"
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

func (i *InviteStore) FindByID(id uuid.UUID) (*store.Invite, error) {
	var invite store.Invite
	err := i.db.Where("id = ?", id).First(&invite).Error
	return &invite, err
}

func (i *InviteStore) DeleteByID(id uuid.UUID) error {
	return i.db.Delete(&store.Invite{}, "id = ?", id).Error
}
