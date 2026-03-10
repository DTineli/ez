package store

import (
	"net/http"
)

type Pagination struct {
	Page       int
	PerPage    int
	TotalPages int
}

type FindResults[T any] struct {
	Count   int64
	Results []T
}

type ListResults[T any] struct {
	Pagination
	Results FindResults[T]
}

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

type Tenant struct {
	ID       uint
	Slug     string `gorm:"uniqueIndex" json:"slug"`
	Document string `json:"document"`
	Users    []User
}

type AccessType string

const (
	AccessAdmin    AccessType = "admin"
	AccessCustomer AccessType = "customer"
)

type Session struct {
	UserAccessType AccessType
	UserID         uint `json:"user_id"`
	UserName       string
	UserEmail      string
	TenantID       uint   `json:"tenant_id"`
	TenantSlug     string `json:"tenant_slug"`
}

type SessionStore interface {
	CreateSession(*http.Request, http.ResponseWriter, Session) error
	DeleteSession(*http.Request, http.ResponseWriter) error
	GetSessionInfo(*http.Request) (*Session, error)
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
