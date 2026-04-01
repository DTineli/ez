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
