package dbstore

import (
	"github.com/DTineli/ez/internal/store"
	"gorm.io/gorm"
)

type UserStore struct {
	db *gorm.DB
}

func NewUserStore(db *gorm.DB) *UserStore {
	return &UserStore{
		db: db,
	}
}

func (u *UserStore) CreateUser(user store.UserDTO) (int, error) {
	return 0, nil
}

func (u *UserStore) GetUser(email string) (*store.User, error) {

	return nil, nil
}
