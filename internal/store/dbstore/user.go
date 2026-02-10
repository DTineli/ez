package dbstore

import (
	"github.com/DTineli/ez/internal/store"
	"golang.org/x/crypto/bcrypt"
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

func (u *UserStore) CreateUser(dto store.UserDTO) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user := store.User{
		Name:     dto.Name,
		Email:    dto.Email,
		Password: string(hashed),
	}
	return u.db.Create(&user).Error
}

func (u *UserStore) GetUser(email string) (*store.User, error) {
	var user store.User
	err := u.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
