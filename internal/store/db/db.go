package db

import (
	"time"

	"github.com/DTineli/ez/internal/store"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func MustOpen(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db
}

func MustMigrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&store.Tenant{},
		&store.User{},
		&store.Product{},
		&store.ProductPrice{},
		&store.PriceTable{},

		&store.Contact{},
		&store.Invite{},
		&store.Cart{},
		&store.CartItem{},
		&store.Order{},
		&store.OrderItem{},
		&store.PaymentMethod{},

		&store.Attribute{},
		&store.AttributeValue{},
		&store.VariantAttribute{},
		&store.Variant{},
	)
	if err != nil {
		panic(err)
	}
}
