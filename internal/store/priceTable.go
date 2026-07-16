package store

type PriceTable struct {
	ID         uint    `gorm:"primaryKey"                                                        json:"id"`
	Name       string  `gorm:"type:varchar(100);not null;uniqueIndex:idx_tenant_name,priority:2" json:"name"`
	Percentage float64 `gorm:"type:decimal(10,2)"`
	Status     bool    `gorm:"default:true"                                                      json:"active"`

	TenantID uint           `gorm:"not null;uniqueIndex:idx_tenant_name,priority:1" json:"tenant_id"`
	Prices   []ProductPrice `gorm:"foreignKey:PriceTableID"                         json:"prices,omitempty"`

	PaymentMethods []PaymentMethod `gorm:"many2many:price_table_payment_methods;" json:"payment_methods,omitempty"`
}

type ProductPrice struct {
	ID    uint
	Price float64

	VariantID    uint       `gorm:"not null;uniqueIndex:idx_variant_pricetable,unique" json:"variant_id"`
	PriceTableID uint       `gorm:"not null;uniqueIndex:idx_variant_pricetable,unique" json:"price_table_id"`
	Variant      Variant    `gorm:"foreignKey:VariantID"                               json:"variant"`
	PriceTable   PriceTable `gorm:"foreignKey:PriceTableID"                            json:"price_table"`
}

type PriceTableStore interface {
	CreatePriceTable(*PriceTable) error
	FindAllByTenant(id uint) ([]PriceTable, error)
	FindAllActiveByTenant(id uint) ([]PriceTable, error)
	FindAllActiveByTenantAndClient(
		tenantID, clientID uint,
	) ([]PriceTable, error)
	GetOne(id uint, tenantID uint) (*PriceTable, error)
	GetOneWithPrices(id, tenantID uint) (*PriceTable, error)
	HasContacts(priceTableID, tenantID uint) (bool, error)
	Delete(id, tenantID uint) error

	CreateProductPrice(*ProductPrice) error
	FindProductPrices(productID uint) ([]ProductPrice, error)
	GetOneProductPrice(id uint) (*ProductPrice, error)

	GetOneProductPriceWithVariant(id uint) (*ProductPrice, error)
	UpdateProductPrice(id uint, price float64) error
	DeleteProductPrice(PriceID uint) error
	SearchVariantsForPriceTable(
		tenantID, priceTableID uint,
		q string,
	) ([]Variant, error)
	FindPriceTablesByProduct(productID, tenantID uint) ([]PriceTable, error)
	FindProductPricesForProduct(productID uint) ([]ProductPrice, error)

	FindPaymentMethods(tableID, tenantID uint) ([]PaymentMethod, error)
}

type VariantTableRow struct {
	Variant Variant
	Price   *ProductPrice
}

type PriceTableProductView struct {
	Table           PriceTable
	Rows            []VariantTableRow
	MissingVariants []Variant
}
