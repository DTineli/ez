package store

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Invite struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Document string    `gorm:"not null;unique"`
	Phone    string

	OriginTenant uint
	Tenant       Tenant `gorm:"foreignKey:OriginTenant"`

	UserID *uint
	User   *User `gorm:"foreignKey:UserID"`
}

func (i *Invite) BeforeCreate(tx *gorm.DB) error {
	i.ID = uuid.New()
	return nil
}

type InviteStore interface {
	Create(*Invite) error
	// Find(filters any) ([]Invite, error)
	// FindByID(id uuid.UUID) (Invite, error)
}
