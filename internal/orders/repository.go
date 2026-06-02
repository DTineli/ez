package orders

import "github.com/DTineli/ez/internal/store"

type Repository interface {
	ConfirmFromCart(cartID, tenantID, contactID, priceTableID uint) (*store.Order, error)
	ListByTenant(tenantID uint) ([]store.AdminOrderListItem, error)
	ListByTenantPaged(tenantID uint, filters store.OrderFilters) ([]store.AdminOrderListItem, int64, error)
	ListByContact(tenantID, contactID uint) ([]store.ClientOrderListItem, error)
	GetByID(id, tenantID uint) (*store.OrderDetail, error)
	Create(tenantID, contactID uint, items []store.NewOrderItem) (*store.Order, error)
	Salvar(order *store.OrderDetail) error
}
