package store

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `json:"name"`
	Email    string `gorm:"uniqueIndex" json:"email"`
	Password string `json:"-"`
	TenantID uint   `json:"tenant_id"`
	Tenant   Tenant
}

type Tenant struct {
	ID       uint
	Slug     string `gorm:"uniqueIndex" json:"slug"`
	Document string `json:"document"`
	Users    []User
}

type Product struct {
	ID       uint    `gorm:"primaryKey" json:"id"`
	SKU      string  `json:"sku"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Stock    int     `json:"stock"`
	Tenant   Tenant
	TenantID uint `json:"owner_id"`
}

type Session struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	SessionID  string `json:"session_id"`
	UserID     uint   `json:"user_id"`
	User       User   `gorm:"foreignKey:UserID" json:"user"`
	TenantID   uint   `json:"tenant_id"`
	TenantSlug string `json:"tenant_slug"`
}

type TenantStore interface {
	CreateTenant(Tenant) (uint, error)

	// GetTenantById(id uint) (*Tenant, error)
	// GetTenantBySlug(slug string) (*Tenant, error)
}

type UserStore interface {
	CreateUser(User) error
	GetUser(email string) (*User, error)
}

type SessionStore interface {
	CreateSession(session *Session) (*Session, error)
	GetUserFromSession(sessionID string, userID string) (*User, error)
}

type ProductStore interface {
	CreateProduct(*Product) error
	GetProduct(id uint) (*Product, error)

	FindAllByUser(userID uint) ([]Product, error)
}
