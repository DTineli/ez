package store

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `json:"name"`
	Email    string `gorm:"uniqueIndex" json:"email"`
	Password string `json:"-"`
}

type Product struct {
	ID     uint    `gorm:"primaryKey" json:"id"`
	SKU    string  `json:"sku"`
	Name   string  `json:"name"`
	Price  float64 `json:"price"`
	Stock  int     `json:"stock"`
	User   User
	UserID uint `json:"owner_id"`
}

type UserDTO struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

type Session struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	SessionID string `json:"session_id"`
	UserID    uint   `json:"user_id"`
	User      User   `gorm:"foreignKey:UserID" json:"user"`
}

type UserStore interface {
	CreateUser(UserDTO) error
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
