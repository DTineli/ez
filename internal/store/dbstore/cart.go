package dbstore

import (
	"errors"

	"github.com/DTineli/ez/internal/store"
	"gorm.io/gorm"
)

type CartStore struct {
	db *gorm.DB
}

func NewCartStore(db *gorm.DB) *CartStore {
	return &CartStore{db: db}
}

func (c *CartStore) FindOpenByID(id, tenantID, contactID uint) (*store.Cart, error) {
	var cart store.Cart
	err := c.db.
		Where("id = ? AND tenant_id = ? AND contact_id = ? AND status = ?", id, tenantID, contactID, store.CartStatusOpen).
		First(&cart).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

func (c *CartStore) FindOpenByContact(tenantID, contactID uint) (*store.Cart, error) {
	var cart store.Cart
	err := c.db.
		Where("tenant_id = ? AND contact_id = ? AND status = ?", tenantID, contactID, store.CartStatusOpen).
		First(&cart).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

func (c *CartStore) Create(cart *store.Cart) error {
	return c.db.Create(cart).Error
}

func (c *CartStore) AddOrIncrementItem(cartID, productID uint, quantity int, unitPrice float64) error {
	return c.db.Transaction(func(tx *gorm.DB) error {
		var item store.CartItem
		err := tx.Where("cart_id = ? AND product_id = ?", cartID, productID).First(&item).Error
		if err == nil {
			return tx.Model(&store.CartItem{}).
				Where("id = ?", item.ID).
				Update("quantity", item.Quantity+quantity).Error
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		return tx.Create(&store.CartItem{
			CartID:    cartID,
			ProductID: productID,
			Quantity:  quantity,
			UnitPrice: unitPrice,
		}).Error
	})
}

func (c *CartStore) CountItems(cartID uint) (int64, error) {
	var total int64
	err := c.db.
		Model(&store.CartItem{}).
		Where("cart_id = ?", cartID).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (c *CartStore) ListCheckoutItems(cartID, tenantID uint) ([]store.CartCheckoutItem, error) {
	type checkoutRow struct {
		ID        uint
		ProductID uint
		Name      string
		Quantity  int
		UnitPrice float64
	}

	var rows []checkoutRow
	err := c.db.
		Table("cart_items ci").
		Select("ci.id, ci.product_id, p.name, ci.quantity, ci.unit_price").
		Joins("JOIN products p ON p.id = ci.product_id").
		Where("ci.cart_id = ? AND p.tenant_id = ?", cartID, tenantID).
		Order("ci.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]store.CartCheckoutItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, store.CartCheckoutItem{
			CartItemID: row.ID,
			ProductID:  row.ProductID,
			Name:       row.Name,
			Quantity:   row.Quantity,
			UnitPrice:  row.UnitPrice,
			Subtotal:   float64(row.Quantity) * row.UnitPrice,
		})
	}

	return items, nil
}

func (c *CartStore) RemoveItem(cartID, productID uint) error {
	return c.db.
		Where("cart_id = ? AND product_id = ?", cartID, productID).
		Delete(&store.CartItem{}).Error
}

func (c *CartStore) UpdateItemQty(cartID, productID uint, quantity int) error {
	return c.db.Model(&store.CartItem{}).
		Where("cart_id = ? AND product_id = ?", cartID, productID).
		Update("quantity", quantity).Error
}
