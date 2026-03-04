package handlers

import (
	"net/http"

	"github.com/DTineli/ez/internal/forms"
	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
)

type ContactHandler struct {
	store store.ContactStore
}

func NewContactHandler(db store.ContactStore) *ContactHandler {
	return &ContactHandler{
		store: db,
	}
}

func (c ContactHandler) GetContactsPage(w http.ResponseWriter, r *http.Request) {

}

func (c ContactHandler) GetContactsForm(w http.ResponseWriter, r *http.Request) {
	Render(templates.ContactForm(forms.New(nil), false), r, w)
}

func (c ContactHandler) PostNewContact(w http.ResponseWriter, r *http.Request) {

}
