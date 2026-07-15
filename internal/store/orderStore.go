package store

import "time"

type OrderStatus string
type OrderAtor string

const (
	OrderPendente  OrderStatus = "pendente"
	OrderAprovado  OrderStatus = "aprovado"
	OrderCompleto  OrderStatus = "completo"
	OrderCancelado OrderStatus = "cancelado"

	OrderEmSeparacao        OrderStatus = "em_separacao"
	OrderEntregue           OrderStatus = "entregue"
	OrderAguardandoRetirada OrderStatus = "aguardando_retirada"
)

const (
	OrderPagamentoPendente OrderStatus = "pagamento_pendente"
	OrderPago              OrderStatus = "pago"
)

const (
	OrderStatusConfirmed OrderStatus = "confirmed"
)

const (
	OrderAtorSeller  OrderAtor = "seller"
	OrderAtorBuyer   OrderAtor = "buyer"
	OrderAtorSistema OrderAtor = "sistema"
)

type OrderTransicao struct {
	De   OrderStatus
	Para OrderStatus
	Ator OrderAtor
}

var transicoesValidas = []OrderTransicao{
	{OrderPendente, OrderAprovado, OrderAtorSeller},
	{OrderPendente, OrderCancelado, OrderAtorBuyer},
	{OrderPendente, OrderCancelado, OrderAtorSeller},

	{OrderAprovado, OrderEmSeparacao, OrderAtorSeller},
	{OrderAprovado, OrderCancelado, OrderAtorSeller},

	{OrderEmSeparacao, OrderCompleto, OrderAtorSeller},
	{OrderEmSeparacao, OrderCancelado, OrderAtorSeller},
	{OrderEmSeparacao, OrderAguardandoRetirada, OrderAtorSeller},

	{OrderAguardandoRetirada, OrderEntregue, OrderAtorSeller},

	{OrderEntregue, OrderCompleto, OrderAtorSistema},
	{OrderAprovado, OrderCompleto, OrderAtorSistema},
}

func PodeTransicionarOrder(de, para OrderStatus, ator OrderAtor) bool {
	for _, t := range transicoesValidas {
		if t.De == de && t.Para == para && t.Ator == ator {
			return true
		}
	}
	return false
}

type Order struct {
	ID            uint        `gorm:"primaryKey"                                                 json:"id"`
	TenantID      uint        `gorm:"not null;index"                                             json:"tenant_id"`
	ContactID     uint        `gorm:"not null;index"                                             json:"contact_id"`
	Status        OrderStatus `gorm:"type:varchar(30);not null;index"                            json:"status"`
	PaymentStatus OrderStatus `gorm:"type:varchar(30);not null;index;default:pagamento_pendente" json:"payment_status"`
	PaymentDate   *time.Time  `                                                                  json:"payment_date"`

	PriceTableId uint       `gorm:"index"                                                      json:"price_table_id"`
	PriceTable   PriceTable `gorm:"foreignKey:PriceTableId"                                    json:"-"`

	TotalAmount float64     `gorm:"not null"           json:"total_amount"`
	CreatedAt   time.Time   `                          json:"created_at"`
	EntregueEm  *time.Time  `                          json:"entregue_em"`
	CanceladoEm *time.Time  `                          json:"cancelado_em"`
	Items       []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
}

type OrderItem struct {
	ID        uint    `gorm:"primaryKey"           json:"id"`
	OrderID   uint    `gorm:"not null;index"       json:"order_id"`
	ProductID uint    `gorm:"not null"             json:"product_id"`
	VariantID uint    `gorm:"not null;default:0"   json:"variant_id"`
	Name      string  `gorm:"not null"             json:"name"`
	Quantity  int     `gorm:"not null"             json:"quantity"`
	UnitPrice float64 `gorm:"not null"             json:"unit_price"`
	Subtotal  float64 `gorm:"not null"             json:"subtotal"`
	Variant   Variant `gorm:"foreignKey:VariantID"`
}

type PaymentMethod struct {
	ID        uint   `gorm:"primaryKey"                                                          json:"id"`
	Name      string `gorm:"type:varchar(100);not null;uniqueIndex:idx_tenant_pm_name,priority:2" json:"name"`
	Ativo     bool   `gorm:"default:true"                                                        json:"active"`
	CanDivide bool   `gorm:"default:false"                                                       json:"can_divide"`
	TenantID  uint   `gorm:"not null;uniqueIndex:idx_tenant_pm_name,priority:1"                  json:"tenant_id"`

	PriceTables []PriceTable `gorm:"many2many:price_table_payment_methods;" json:"price_tables,omitempty"`
}

type PaymentMethodStore interface {
	CreatePaymentMethod(pm *PaymentMethod) error
	GetPaymentMethod(id, tenantID uint) (*PaymentMethod, error)
	FindAllPaymentMethodsByTenant(tenantID uint) ([]PaymentMethod, error)
	UpdatePaymentMethod(pm *PaymentMethod) error
	DeletePaymentMethod(id, tenantID uint) error
}

type AdminOrderListItem struct {
	ID          uint
	ContactName string
	TradeName   string
	Status      OrderStatus
	TotalAmount float64
	CreatedAt   time.Time
}

type OrderDetail struct {
	ID            uint
	ContactID     uint
	ContactName   string
	Status        OrderStatus
	PaymentStatus OrderStatus
	PaymentDate   *time.Time
	TotalAmount   float64
	CreatedAt     time.Time
	EntregueEm    *time.Time
	CanceladoEm   *time.Time
	Items         []OrderItem
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

type OrderFilters struct {
	Page        int
	PerPage     int
	ContactName string
	Status      OrderStatus
}

type AdminOrderListPage struct {
	Orders     []AdminOrderListItem
	Filters    OrderFilters
	TotalPages int
	Total      int64
}
