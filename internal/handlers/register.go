package handlers

import (
	"net/http"

	"gorm.io/gorm"
)

type RegisterService struct {
	db *gorm.DB
}

type RegisterHandler struct {
	service *RegisterService
}

func NewRegisterHandlerWithService(db *gorm.DB) *RegisterHandler {
	return &RegisterHandler{
		service: &RegisterService{
			db: db,
		},
	}
}

func (l *RegisterService) GetRegisterPage(w http.ResponseWriter, r *http.Request) {

}
