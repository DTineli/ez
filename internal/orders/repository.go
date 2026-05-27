package orders

type Repository interface {
	ConfirmFromCart(cartID, tenantID, contactID, priceTableID uint) (*Order, error)
	ListByTenant(tenantID uint) ([]AdminOrderListItem, error)
	ListByTenantPaged(tenantID uint, filters OrderFilters) ([]AdminOrderListItem, int64, error)
	ListByContact(tenantID, contactID uint) ([]ClientOrderListItem, error)
	GetByID(id, tenantID uint) (*OrderDetail, error)
	Create(tenantID, contactID uint, items []NewOrderItem) (*Order, error)
	Salvar(order *OrderDetail) error
}
