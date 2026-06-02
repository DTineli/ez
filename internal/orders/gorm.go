package orders

import (
	"errors"

	"github.com/DTineli/ez/internal/store"
	"gorm.io/gorm"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (o *GormRepository) ConfirmFromCart(
	cartID, tenantID, contactID, priceTableID uint,
) (*store.Order, error) {
	var created store.Order

	err := o.db.Transaction(func(tx *gorm.DB) error {
		var cart store.Cart

		if err := tx.Where(
			"id = ? AND tenant_id = ? AND contact_id = ? AND status = ?",
			cartID,
			tenantID,
			contactID,
			store.CartStatusOpen,
		).First(&cart).Error; err != nil {
			return err
		}

		var cartItems []store.CartItem
		if err := tx.Where(
			"cart_id = ?",
			cartID,
		).Find(&cartItems).Error; err != nil {
			return err
		}
		if len(cartItems) == 0 {
			return errors.New("cart is empty")
		}

		variantIDs := make([]uint, 0, len(cartItems))
		for _, item := range cartItems {
			variantIDs = append(variantIDs, item.VariantID)
		}

		var variants []store.Variant
		if err := tx.Preload("Product").Where(
			"tenant_id = ? AND id IN ?",
			tenantID,
			variantIDs,
		).Find(&variants).Error; err != nil {
			return err
		}

		productNameByVariantID := make(map[uint]string, len(variants))
		for _, p := range variants {
			productNameByVariantID[p.ID] = p.Product.Name
		}

		var priceTable *store.PriceTable
		if priceTableID != 0 {
			var pt store.PriceTable
			if err := tx.Where("id = ? AND tenant_id = ?", priceTableID, tenantID).First(&pt).Error; err == nil {
				priceTable = &pt
			}
		}

		total := 0.0
		orderItems := make([]store.OrderItem, 0, len(cartItems))
		for _, item := range cartItems {
			name := productNameByVariantID[item.VariantID]
			if name == "" {
				return errors.New("product not found for cart item")
			}

			unitPrice := item.CostPrice
			if priceTable != nil {
				unitPrice = item.CostPrice * (1 + priceTable.Percentage/100)
			}
			subtotal := float64(item.Quantity) * unitPrice
			total += subtotal

			orderItems = append(orderItems, store.OrderItem{
				ProductID: item.ProductID,
				VariantID: item.VariantID,
				Name:      name,
				Quantity:  item.Quantity,
				UnitPrice: unitPrice,
				Subtotal:  subtotal,
			})
		}

		order := store.Order{
			TenantID:    tenantID,
			ContactID:   contactID,
			Status:      store.OrderPendente,
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

func (o *GormRepository) ListByTenant(
	tenantID uint,
) ([]store.AdminOrderListItem, error) {
	var rows []store.AdminOrderListItem
	err := o.db.Table("orders o").
		Select("o.id, c.name as contact_name, o.status, o.total_amount, o.created_at").
		Joins("JOIN contacts c ON c.id = o.contact_id").
		Where("o.tenant_id = ?", tenantID).
		Order("o.id DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (o *GormRepository) ListByTenantPaged(
	tenantID uint,
	filters store.OrderFilters,
) ([]store.AdminOrderListItem, int64, error) {
	var rows []store.AdminOrderListItem
	var count int64

	q := o.db.Table("orders o").
		Joins("JOIN contacts c ON c.id = o.contact_id").
		Where("o.tenant_id = ?", tenantID)

	if filters.ContactName != "" {
		q = q.Where("c.name LIKE ?", "%"+filters.ContactName+"%")
	}
	if filters.Status != "" {
		q = q.Where("o.status = ?", filters.Status)
	}

	if err := q.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	offset := (filters.Page - 1) * filters.PerPage
	err := q.Select("o.id, c.name as contact_name, o.status, o.total_amount, o.created_at").
		Order("o.id DESC").
		Offset(offset).
		Limit(filters.PerPage).
		Scan(&rows).Error

	return rows, count, err
}

func (o *GormRepository) ListByContact(
	tenantID, contactID uint,
) ([]store.ClientOrderListItem, error) {
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

func (o *GormRepository) GetByID(id, tenantID uint) (*store.OrderDetail, error) {
	var order store.Order
	if err := o.db.Preload("Items.Variant.Attributes.AttributeValue").Where(
		"id = ? AND tenant_id = ?",
		id,
		tenantID,
	).First(&order).Error; err != nil {
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
		EntregueEm:  order.EntregueEm,
		CanceladoEm: order.CanceladoEm,
		Items:       order.Items,
	}, nil
}

func (o *GormRepository) Create(
	tenantID, contactID uint,
	items []store.NewOrderItem,
) (*store.Order, error) {
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
		if err := tx.Where("tenant_id = ? AND id IN ?",
			tenantID,
			productIDs,
		).Joins("variants").Find(&products).Error; err != nil {
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
			Status:      store.OrderPendente,
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

func (o *GormRepository) Salvar(order *store.OrderDetail) error {
	updates := map[string]any{
		"status": order.Status,
	}
	if order.EntregueEm != nil {
		updates["entregue_em"] = order.EntregueEm
	}
	if order.CanceladoEm != nil {
		updates["cancelado_em"] = order.CanceladoEm
	}
	return o.db.Model(&store.Order{}).Where("id = ?", order.ID).Updates(updates).Error
}
