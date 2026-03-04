package handlers

import "github.com/DTineli/ez/internal/store"

type ContactHandler struct {
	store store.ContactStore
}

func NewContactHandler(db store.ContactStore) *ContactHandler {
	return &ContactHandler{
		store: db,
	}
}
