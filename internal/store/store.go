package store

import (
	"net/http"
)

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

type Session struct {
	UserID     uint `json:"user_id"`
	UserName   string
	UserEmail  string
	TenantID   uint   `json:"tenant_id"`
	TenantSlug string `json:"tenant_slug"`
}

type TenantStore interface {
	CreateTenant(Tenant) (uint, error)

	GetTenantByID(id uint) (*Tenant, error)
	// GetTenantBySlug(slug string) (*Tenant, error)
}

type UserStore interface {
	CreateUser(User) error
	GetUser(email string) (*User, error)

	// GetUserById(id uint) (*User, error)
}

type SessionStore interface {
	CreateSession(*http.Request, http.ResponseWriter, Session) error
	GetSessionInfo(*http.Request) (*Session, error)
}
