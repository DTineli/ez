package db

import (
	"os"

	"github.com/DTineli/ez/internal/store"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func open(dbName string) (*gorm.DB, error) {

	// make the temp directory if it doesn't exist
	err := os.MkdirAll("/tmp", 0755)
	if err != nil {
		return nil, err
	}

	return gorm.Open(sqlite.Open(dbName), &gorm.Config{})
}

func MustOpen(dbName string) *gorm.DB {

	if dbName == "" {
		dbName = "goth.db"
	}

	db, err := open(dbName)
	if err != nil {
		panic(err)
	}

	// Drop stale cart data before migrating CartItem schema (variant_id index change)
	db.Exec("DELETE FROM cart_items WHERE variant_id = 0 OR variant_id IS NULL")
	db.Exec("DROP INDEX IF EXISTS idx_cart_product")
	// SQLite cannot add NOT NULL column to existing rows with NULL value
	db.Exec(
		"DELETE FROM order_items WHERE variant_id IS NULL OR variant_id = 0",
	)
	// Remove orders with no items
	db.Exec(
		"DELETE FROM orders WHERE id NOT IN (SELECT DISTINCT order_id FROM order_items)",
	)

	err = db.AutoMigrate(
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

		&store.Attribute{},
		&store.AttributeValue{},
		&store.VariantAttribute{},
		&store.Variant{},
	)

	if err != nil {
		panic(err)
	}

	return db
}
