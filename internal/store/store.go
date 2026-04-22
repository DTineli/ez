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

type Tenant struct {
	ID       uint
	Slug     string `gorm:"uniqueIndex" json:"slug"`
	Document string `json:"document"`
	Users    []User
}

type ContactInfo struct {
	ID         uint
	PriceTable uint
}

const (
	AdminSessionName  = "ez_admin_session"
	ClientSessionName = "ez_client_session"
)

type Session struct {
	Name           string
	UserAccessType AccessType
	UserID         uint `json:"user_id"`
	UserName       string
	UserEmail      string
	TenantID       uint   `json:"tenant_id"`
	TenantSlug     string `json:"tenant_slug"`
	CartID         uint   `json:"cart_id"`

	ContactInfo *ContactInfo
}

type SessionStore interface {
	CreateSession(*http.Request, http.ResponseWriter, Session) error
	DeleteSession(*http.Request, http.ResponseWriter) error
	GetSessionInfo(*http.Request) (*Session, error)
	SetCartID(*http.Request, http.ResponseWriter, uint) error
}

type TenantStore interface {
	CreateTenant(Tenant) (uint, error)

	GetTenantByID(id uint) (*Tenant, error)
	GetTenantBySlug(slug string) (*Tenant, error)
}
