package store

type CartStatus string

const (
	CartStatusOpen CartStatus = "open"
)

type Cart struct {
	ID        uint       `gorm:"primaryKey"                                                                json:"id"`
	TenantID  uint       `gorm:"not null;index:idx_cart_tenant_contact_status,priority:1"                  json:"tenant_id"`
	ContactID uint       `gorm:"not null;index:idx_cart_tenant_contact_status,priority:2"                  json:"contact_id"`
	Status    CartStatus `gorm:"type:varchar(20);not null;index:idx_cart_tenant_contact_status,priority:3" json:"status"`
	Items     []CartItem `gorm:"foreignKey:CartID"                                                         json:"items"`
}

type CartItem struct {
	ID        uint    `gorm:"primaryKey"                                                 json:"id"`
	CartID    uint    `gorm:"not null;uniqueIndex:idx_cart_variant,priority:1"           json:"cart_id"`
	VariantID uint    `gorm:"not null;default:0;uniqueIndex:idx_cart_variant,priority:2" json:"variant_id"`
	ProductID uint    `gorm:"not null"                                                   json:"product_id"`
	Quantity  int     `gorm:"not null"                                                   json:"quantity"`
	UnitPrice float64 `gorm:"not null"                                                   json:"unit_price"`
}

type CartCheckoutItem struct {
	CartItemID   uint
	ProductID    uint
	Name         string
	VariantLabel string
	VariantID    uint
	Quantity     int
	UnitPrice    float64
	Subtotal     float64
}

type CartStore interface {
	FindOpenByID(id, tenantID, contactID uint) (*Cart, error)
	FindOpenByContact(tenantID, contactID uint) (*Cart, error)
	Create(*Cart) error
	AddOrIncrementItem(
		cartID, productID,
		variantID uint,
		quantity int,
		unitPrice float64,
	) error
	CountItems(cartID uint) (int64, error)
	ListCheckoutItems(cartID, tenantID uint) ([]CartCheckoutItem, error)
	RemoveItem(cartID, productID uint) error
	UpdateItemQty(cartID, productID uint, variantID uint, quantity int) error
}
