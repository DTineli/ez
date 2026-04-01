package store

type AccessType string

const (
	AccessAdmin    AccessType = "admin"
	AccessCustomer AccessType = "customer"
)

type User struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserAccess AccessType `gorm:"varchar(20)"`
	Name       string     `json:"name"`
	Email      string     `gorm:"uniqueIndex" json:"email"`
	Phone      string     `gorm:"uniqueIndex:idx_tenant_phone"`
	Password   string     `json:"-"`
	TenantID   uint       `gorm:"uniqueIndex:idx_tenant_phone"`
	Tenant     Tenant
}

type UserStore interface {
	CreateUser(User) error
	GetUser(email string) (*User, error)

	// GetUserById(id uint) (*User, error)
}
