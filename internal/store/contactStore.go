package store

type ContactType string
type DocumentType string

const (
	Customer ContactType = "customer"
	Supplier ContactType = "supplier"
)

const (
	PFisica   DocumentType = "p_fisica"
	PJuridica DocumentType = "p_juridica"
)

type ContactFilters struct {
	Pagination

	Name        string `filter:"name,like"`
	TradeName   string `filter:"trade_name,like"`
	Document    string `filter:"document,eq"`
	ContactType string `filter:"contact_type,eq"`
}

type Contact struct {
	ID          uint        `gorm:"primaryKey" json:"id"`
	Name        string      `json:"name"`
	TradeName   string      `json:"trade_name"`
	ContactType ContactType `json:"contact_type"`

	DocumentType string `gorm:"type:varchar(12);default:'p_juridica'" json:"document_type"`
	Document     string `gorm:"type:varchar(50);not null;index:idx_tenant_document,unique,priority:2" json:"document"`
	IE           string `gorm:"type:varchar(20)" json:"ie"`

	Email string `json:"email"`
	Phone string `json:"phone"`

	// Endereço completo
	ZipCode      string `gorm:"type:varchar(20)" json:"zipcode"`
	Street       string `gorm:"type:varchar(100)" json:"street"`
	Number       string `gorm:"type:varchar(20)" json:"number"`
	Complement   string `gorm:"type:varchar(50)" json:"complement"`
	Neighborhood string `gorm:"type:varchar(50)" json:"neighborhood"`
	City         string `gorm:"type:varchar(50)" json:"city"`
	UF           string `gorm:"type:varchar(2)" json:"uf"` // Estado sigla

	PriceTableID uint `gorm:"type:int" json:"price_table"`
	TenantID     uint `gorm:"not null;index:idx_tenant_document,priority:1" json:"tenant_id"`
}

type ContactStore interface {
	CreateContact(*Contact) error
	FindAll(uint, ContactFilters) (*FindResults[Contact], error)
	GetOne(uint) (*Contact, error)
	UpdateById(id, tenantID uint, fields map[string]any) error
	GetOneByPhone(phone string) (*Contact, error)
}
