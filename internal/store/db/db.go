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
	// Migrate unit_price → cost_price in cart_items
	var hasUnitPrice int
	db.Raw("SELECT COUNT(*) FROM pragma_table_info('cart_items') WHERE name = 'unit_price'").Scan(&hasUnitPrice)
	if hasUnitPrice > 0 {
		db.Exec("ALTER TABLE cart_items ADD COLUMN cost_price real NOT NULL DEFAULT 0")
		db.Exec("UPDATE cart_items SET cost_price = unit_price")
		db.Exec("CREATE TABLE cart_items_new AS SELECT id, cart_id, variant_id, product_id, quantity, cost_price FROM cart_items")
		db.Exec("DROP TABLE cart_items")
		db.Exec("ALTER TABLE cart_items_new RENAME TO cart_items")
	}
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
