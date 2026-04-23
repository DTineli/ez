package dbstore

import (
	"errors"

	"github.com/DTineli/ez/internal/store"
	"gorm.io/gorm"
)

type OrderStore struct {
	db *gorm.DB
}

func NewOrderStore(db *gorm.DB) *OrderStore {
	return &OrderStore{db: db}
}

func (o *OrderStore) ConfirmFromCart(cartID, tenantID, contactID uint) (*store.Order, error) {
	var created store.Order

	err := o.db.Transaction(func(tx *gorm.DB) error {
		var cart store.Cart
		if err := tx.Where("id = ? AND tenant_id = ? AND contact_id = ? AND status = ?", cartID, tenantID, contactID, store.CartStatusOpen).First(&cart).Error; err != nil {
			return err
		}

		var cartItems []store.CartItem
		if err := tx.Where("cart_id = ?", cartID).Find(&cartItems).Error; err != nil {
			return err
		}
		if len(cartItems) == 0 {
			return errors.New("cart is empty")
		}

		productIDs := make([]uint, 0, len(cartItems))
		for _, item := range cartItems {
			productIDs = append(productIDs, item.ProductID)
		}

		var products []store.Product
		if err := tx.Where("tenant_id = ? AND id IN ?", tenantID, productIDs).Find(&products).Error; err != nil {
			return err
		}

		productNameByID := make(map[uint]string, len(products))
		for _, p := range products {
			productNameByID[p.ID] = p.Name
		}

		total := 0.0
		orderItems := make([]store.OrderItem, 0, len(cartItems))
		for _, item := range cartItems {
			name := productNameByID[item.ProductID]
			if name == "" {
				return errors.New("product not found for cart item")
			}

			subtotal := float64(item.Quantity) * item.UnitPrice
			total += subtotal

			orderItems = append(orderItems, store.OrderItem{
				ProductID: item.ProductID,
				Name:      name,
				Quantity:  item.Quantity,
				UnitPrice: item.UnitPrice,
				Subtotal:  subtotal,
			})
		}

		order := store.Order{
			TenantID:    tenantID,
			ContactID:   contactID,
			Status:      store.OrderStatusConfirmed,
			TotalAmount: total,
			Items:       orderItems,
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		if err := tx.Model(&store.Cart{}).Where("id = ?", cartID).Update("status", "confirmed").Error; err != nil {
			return err
		}

		created = order
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (o *OrderStore) ListByTenant(tenantID uint) ([]store.AdminOrderListItem, error) {
	var modelRows []store.AdminOrderListItem
	err := o.db.Table("orders o").
		Select("o.id, c.name as contact_name, o.status, o.total_amount, o.created_at").
		Joins("JOIN contacts c ON c.id = o.contact_id").
		Where("o.tenant_id = ?", tenantID).
		Order("o.id DESC").
		Scan(&modelRows).Error
	if err != nil {
		return nil, err
	}

	return modelRows, nil
}

func (o *OrderStore) ListByContact(tenantID, contactID uint) ([]store.ClientOrderListItem, error) {
	var rows []store.ClientOrderListItem
	err := o.db.Table("orders").
		Select("id, status, total_amount, created_at").
		Where("tenant_id = ? AND contact_id = ?", tenantID, contactID).
		Order("id DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (o *OrderStore) GetByID(id, tenantID uint) (*store.OrderDetail, error) {
	var order store.Order
	if err := o.db.Preload("Items").Where("id = ? AND tenant_id = ?", id, tenantID).First(&order).Error; err != nil {
		return nil, err
	}

	var contact store.Contact
	if err := o.db.Select("name").Where("id = ?", order.ContactID).First(&contact).Error; err != nil {
		return nil, err
	}

	return &store.OrderDetail{
		ID:          order.ID,
		ContactID:   order.ContactID,
		ContactName: contact.Name,
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
		CreatedAt:   order.CreatedAt,
		Items:       order.Items,
	}, nil
}

func (o *OrderStore) Create(tenantID, contactID uint, items []store.NewOrderItem) (*store.Order, error) {
	var created store.Order

	err := o.db.Transaction(func(tx *gorm.DB) error {
		if len(items) == 0 {
			return errors.New("no items")
		}

		productIDs := make([]uint, 0, len(items))
		for _, item := range items {
			productIDs = append(productIDs, item.ProductID)
		}

		var products []store.Product
		if err := tx.Where("tenant_id = ? AND id IN ?", tenantID, productIDs).Find(&products).Error; err != nil {
			return err
		}

		productNameByID := make(map[uint]string, len(products))
		for _, p := range products {
			productNameByID[p.ID] = p.Name
		}

		total := 0.0
		orderItems := make([]store.OrderItem, 0, len(items))
		for _, item := range items {
			name := productNameByID[item.ProductID]
			if name == "" {
				return errors.New("product not found")
			}
			subtotal := float64(item.Quantity) * item.UnitPrice
			total += subtotal
			orderItems = append(orderItems, store.OrderItem{
				ProductID: item.ProductID,
				Name:      name,
				Quantity:  item.Quantity,
				UnitPrice: item.UnitPrice,
				Subtotal:  subtotal,
			})
		}

		order := store.Order{
			TenantID:    tenantID,
			ContactID:   contactID,
			Status:      store.OrderStatusConfirmed,
			TotalAmount: total,
			Items:       orderItems,
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		created = order
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &created, nil
}
