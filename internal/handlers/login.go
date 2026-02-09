package handlers

import (
	"net/http"

	"gorm.io/gorm"
)

type LoginService struct {
	db *gorm.DB
}

type LoginHandler struct {
	service *LoginService
}

func NewLoginHandlerWithService(db *gorm.DB) *LoginHandler {
	return &LoginHandler{
		service: &LoginService{
			db: db,
		},
	}
}

func (l *LoginHandler) GetLoginPage(w http.ResponseWriter, r *http.Request) {

}
