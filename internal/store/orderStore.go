package store

import "time"

type OrderStatus string

const (
	OrderStatusConfirmed OrderStatus = "confirmed"
)

type Order struct {
	ID          uint        `gorm:"primaryKey"                      json:"id"`
	TenantID    uint        `gorm:"not null;index"                  json:"tenant_id"`
	ContactID   uint        `gorm:"not null;index"                  json:"contact_id"`
	Status      OrderStatus `gorm:"type:varchar(20);not null;index" json:"status"`
	TotalAmount float64     `gorm:"not null"                        json:"total_amount"`
	CreatedAt   time.Time   `                                       json:"created_at"`
	Items       []OrderItem `gorm:"foreignKey:OrderID"              json:"items"`
}

type OrderItem struct {
	ID        uint    `gorm:"primaryKey"         json:"id"`
	OrderID   uint    `gorm:"not null;index"     json:"order_id"`
	ProductID uint    `gorm:"not null"           json:"product_id"`
	VariantID uint    `gorm:"not null;default:0" json:"variant_id"`
	Name      string  `gorm:"not null"           json:"name"`
	Quantity  int     `gorm:"not null"           json:"quantity"`
	UnitPrice float64 `gorm:"not null"           json:"unit_price"`
	Subtotal  float64 `gorm:"not null"           json:"subtotal"`

	Variant Variant `gorm:"foreignKey:VariantID"`
}

type AdminOrderListItem struct {
	ID          uint
	ContactName string
	Status      OrderStatus
	TotalAmount float64
	CreatedAt   time.Time
}

type OrderDetail struct {
	ID          uint
	ContactID   uint
	ContactName string
	Status      OrderStatus
	TotalAmount float64
	CreatedAt   time.Time
	Items       []OrderItem
}

type NewOrderItem struct {
	ProductID uint
	VariantID uint
	Quantity  int
	UnitPrice float64
}

type ClientOrderListItem struct {
	ID          uint
	Status      OrderStatus
	TotalAmount float64
	CreatedAt   time.Time
}

type OrderStore interface {
	ConfirmFromCart(cartID, tenantID, contactID uint) (*Order, error)
	ListByTenant(tenantID uint) ([]AdminOrderListItem, error)
	ListByContact(tenantID, contactID uint) ([]ClientOrderListItem, error)
	GetByID(id, tenantID uint) (*OrderDetail, error)
	Create(tenantID, contactID uint, items []NewOrderItem) (*Order, error)
}
