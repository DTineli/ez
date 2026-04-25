package store

func (p Product) StatusToString() string {
	if p.Status {
		return "Ativo"
	}

	return "Inativo"
}

func (p Product) DefaultCostPrice() float64 {
	if len(p.Variants) > 0 {
		return p.Variants[0].CostPrice
	}
	return 0
}

type CardData struct {
	ID         uint
	Name       string
	Price      float64
	Photo_Link string
}

type GetProductPageParams struct {
	Page       int
	PerPage    int
	TotalPages int

	Total int

	Products []Product
}

type UOM string

const (
	Unit     UOM = "UN" // Unidade
	Kilogram UOM = "KG" // Quilograma
	Liter    UOM = "LT" // Litro
	Box      UOM = "CX" // Caixa
	Meter    UOM = "MT" // Metro
)

// Atributo representa uma característica de variação (ex: Cor, Tamanho)
type Attribute struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"type:varchar(100);not null" json:"name"`

	Values   []AttributeValue `gorm:"foreignKey:AttributeID" json:"values,omitempty"`
	TenantID uint             `gorm:"not null" json:"tenant_id"`
}

// AttributeValue representa um valor de um atributo (ex: Vermelho, P, M, G)
type AttributeValue struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Value string `gorm:"type:varchar(100);not null;uniqueIndex:idx_attribute_value,priority:2" json:"value"`

	AttributeID uint      `gorm:"not null;uniqueIndex:idx_attribute_value,priority:1" json:"attribute_id"`
	Attribute   Attribute `gorm:"foreignKey:AttributeID" json:"attribute,omitempty"`
}

// VariantAttribute associa uma Variant a um AttributeValue
type VariantAttribute struct {
	VariantID uint `gorm:"primaryKey" json:"variant_id"`

	AttributeValueID uint           `gorm:"primaryKey" json:"attribute_value_id"`
	AttributeValue   AttributeValue `gorm:"foreignKey:AttributeValueID" json:"attribute_value,omitempty"`
}

type Variant struct {
	ID  uint   `gorm:"primaryKey" json:"id"`
	SKU string `gorm:"type:varchar(50);not null;index:idx_variant_tenant_sku,unique,priority:2" json:"sku"`

	CostPrice    float64 `json:"cost_price"`
	CurrentStock int     `gorm:"default:0" json:"current_stock"`
	MinimumStock int     `gorm:"default:0" json:"minimum_stock"`

	Weight   float64 `gorm:"type:decimal(10,3)" json:"weight"`
	HeightCm float64 `gorm:"type:decimal(10,2)" json:"height_cm"`
	WidthCm  float64 `gorm:"type:decimal(10,2)" json:"width_cm"`
	LengthCm float64 `gorm:"type:decimal(10,2)" json:"length_cm"`

	Attributes []VariantAttribute `gorm:"foreignKey:VariantID" json:"attributes,omitempty"`
	Prices     []ProductPrice     `gorm:"foreignKey:VariantID" json:"prices,omitempty"`

	ProductID uint    `gorm:"not null" json:"product_id"`
	Product   Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`

	TenantID uint `gorm:"not null;index:idx_variant_tenant_sku,priority:1" json:"tenant_id"`
}

type Product struct {
	ID              uint   `gorm:"primaryKey" json:"id"`
	SKU             string `gorm:"type:varchar(50);not null;index:idx_tenant_sku,unique,priority:2" json:"sku"`
	Name            string `json:"name"`
	FullDescription string `gorm:"type:mediumtext" json:"full_description"`
	Status          bool   `json:"status"`

	//TODO: FOTOS ??

	UOM UOM    `gorm:"type:varchar(10);default:'UN'" json:"uom"`
	EAN string `gorm:"type:varchar(20);" json:"ean"`
	NCM string `gorm:"type:varchar(20);" json:"ncm"`

	Variants []Variant `gorm:"foreignKey:ProductID" json:"variants,omitempty"`

	TenantID uint `gorm:"not null;index:idx_tenant_sku,priority:1" json:"tenant_id"`
}

type PriceTable struct {
	ID         uint    `gorm:"primaryKey" json:"id"`
	Name       string  `gorm:"type:varchar(100);not null;uniqueIndex:idx_tenant_name,priority:2" json:"name"`
	Percentage float64 `gorm:"type:decimal(10,2)"`
	Status     bool    `gorm:"default:true" json:"active"`

	TenantID uint           `gorm:"not null;uniqueIndex:idx_tenant_name,priority:1" json:"tenant_id"`
	Prices   []ProductPrice `gorm:"foreignKey:PriceTableID" json:"prices,omitempty"`
}

type ProductPrice struct {
	ID    uint
	Price float64

	VariantID    uint       `gorm:"not null;uniqueIndex:idx_variant_pricetable,unique" json:"variant_id"`
	PriceTableID uint       `gorm:"not null;uniqueIndex:idx_variant_pricetable,unique" json:"price_table_id"`
	Variant      Variant    `gorm:"foreignKey:VariantID" json:"variant"`
	PriceTable   PriceTable `gorm:"foreignKey:PriceTableID" json:"price_table"`
}

type PriceTableStore interface {
	CreatePriceTable(*PriceTable) error
	FindAllByTenant(id uint) ([]PriceTable, error)
	GetOne(id uint, tenantID uint) (*PriceTable, error)
	HasContacts(priceTableID, tenantID uint) (bool, error)
	Delete(id, tenantID uint) error
}

type ProductFilters struct {
	Page    int
	PerPage int
	SKU     string
	Name    string
	Search  string // OR entre name LIKE e sku LIKE
}

type ProductStore interface {
	CreateProduct(*Product) error
	UpdateFields(id uint, tenantID uint, fields map[string]any) error
	GetProduct(id uint) (*Product, error)
	FindAllByUserWithFilters(id uint, filters ProductFilters) (*FindResults[Product], error)
	FindAllByUser(userID uint) ([]Product, error)

	// Variant
	CreateVariant(*Variant) error
	GetVariant(id uint, tenantID uint) (*Variant, error)
	FindVariantsByProduct(productID uint, tenantID uint) ([]Variant, error)
	UpdateVariantFields(id uint, tenantID uint, fields map[string]any) error
	DeleteVariant(id uint, tenantID uint) error
	SetVariantAttributes(variantID uint, attributeValueIDs []uint) error

	// Attribute
	CreateAttribute(*Attribute) error
	GetAttribute(id uint, tenantID uint) (*Attribute, error)
	FindAttributesByTenant(tenantID uint) ([]Attribute, error)
	DeleteAttribute(id uint, tenantID uint) error
	CreateAttributeValue(*AttributeValue) error
	DeleteAttributeValue(id uint, tenantID uint) error
}
