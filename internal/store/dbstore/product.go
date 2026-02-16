package dbstore

import (
	"github.com/DTineli/ez/internal/store"
	"gorm.io/gorm"
)

type ProductStore struct {
	db *gorm.DB
}

func NewProductStore(db *gorm.DB) *ProductStore {
	return &ProductStore{
		db: db,
	}
}

func (p *ProductStore) CreateProduct(user *store.Product) error {
	return p.db.Create(user).Error
}

func (p *ProductStore) GetProduct(id uint) (*store.Product, error) {
	var product store.Product
	err := p.db.Where("id = ?", id).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (p *ProductStore) FindAllByUser(userID uint) ([]store.Product, error) {
	var products []store.Product

	err := p.db.Where("tenant_id = ?", userID).Find(&products).Error
	if err != nil {
		return nil, err
	}

	return products, nil
}
