package store

import (
	"time"

	"gorm.io/gorm"
)

type UOM string

const (
	Unit     UOM = "UN" // Unidade
	Kilogram UOM = "KG" // Quilograma
	Liter    UOM = "LT" // Litro
	Box      UOM = "CX" // Caixa
	Meter    UOM = "MT" // Metro
)

type Product struct {
	ID              uint   `gorm:"primaryKey" json:"id"`
	SKU             string `gorm:"type:varchar(50);not null;index:idx_tenant_sku,unique,priority:2" json:"sku"`
	Name            string `json:"name"`
	FullDescription string `gorm:"type:text" json:"full_description"`
	Status          bool   `json:"status"`

	//TODO: FOTOS ??

	UOM UOM    `gorm:"type:varchar(10);default:'UN'" json:"uom"` //UNIDADE DE MEDIDA
	EAN string `gorm:"type:varchar(20);" json:"ean"`
	NCM string `gorm:"type:varchar(20);" json:"ncm"`

	// Variacao
	IsVariant bool     `gorm:"default:false" json:"is_variant"`
	ParentID  uint     `gorm:"index" json:"parent_id,omitempty"`
	Parent    *Product `gorm:"foreignKey:ParentID"`

	CostPrice    float64 `json:"cost_price"`
	CurrentStock int     `gorm:"default:0" json:"current_stock"`
	MinimumStock int     `gorm:"default:0" json:"minimum_stock"`

	//Dimensoes
	Weight   float64 `gorm:"type:decimal(10,3)" json:"weight"`
	HeightCm float64 `gorm:"type:decimal(10,2)" json:"height_cm"`
	WidthCm  float64 `gorm:"type:decimal(10,2)" json:"width_cm"`
	LengthCm float64 `gorm:"type:decimal(10,2)" json:"length_cm"`

	Prices   []ProductPrice `gorm:"foreignKey:ProductID" json:"prices,omitempty"`
	TenantID uint           `gorm:"not null;index:idx_tenant_sku,priority:1" json:"tenant_id"`
}

type PriceTable struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Name   string `gorm:"type:varchar(100);not null;index" json:"name"`
	Status bool   `gorm:"default:true" json:"active"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Prices []ProductPrice `gorm:"foreignKey:PriceTableID" json:"prices,omitempty"`
}

type ProductPrice struct {
	ID    uint
	Price float64

	ProductID    uint       `gorm:"not null;index:idx_product_pricetable,unique" json:"product_id"`
	PriceTableID uint       `gorm:"not null;index:idx_product_pricetable,unique" json:"price_table_id"`
	Product      Product    `gorm:"foreignKey:ProductID" json:"product"`
	PriceTable   PriceTable `gorm:"foreignKey:PriceTableID" json:"price_table"`
}

type ProductStore interface {
	CreateProduct(*Product) error
	GetProduct(id uint) (*Product, error)

	FindAllByUser(userID uint) ([]Product, error)
}
