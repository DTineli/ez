package store

type ContactType string

const (
	Customer ContactType = "customer" // Unidade
	Supplier ContactType = "supplier" // Quilograma
)

type Contact struct {
	ID          uint        `gorm:"primaryKey" json:"id"`
	Name        string      `json:"name"`
	FantasyName string      `json:"fantasy_name"`
	Document    string      `gorm:"type:varchar(50);not null;index:idx_tenant_document,unique,priority:2" json:"document"`
	ContactType ContactType `json:"contact_type"`

	Email string `json:"email"`
	Phone string `json:"phone"`

	// Endereço completo
	Street       string `gorm:"type:varchar(100)" json:"street"`
	Number       string `gorm:"type:varchar(20)" json:"number"`
	Complement   string `gorm:"type:varchar(50)" json:"complement"`
	Neighborhood string `gorm:"type:varchar(50)" json:"neighborhood"`
	City         string `gorm:"type:varchar(50)" json:"city"`
	UF           string `gorm:"type:varchar(2)" json:"uf"` // Estado sigla
	ZipCode      string `gorm:"type:varchar(20)" json:"zipcode"`

	PriceTableID uint `gorm:"uniqueIndex:idx_contact_pricetable,unique" json:"price_table"`

	TenantID uint `gorm:"not null;index:idx_tenant_document,priority:1" json:"tenant_id"`
}

type ContactStore interface {
	// CreateContact(*Contact) error
}
