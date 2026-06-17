// Package legacy lê a tabela `products` flat do banco SQLite antigo
// (bancoProd.db), que mistura campos de produto e variante num schema
// que não existe mais no código atual.
package legacy

import "database/sql"

// LegacyProduct é uma linha da tabela `products` do bancoProd.db.
// Usa sql.Null* pra distinguir NULL de zero-value, já que vários
// campos têm NULLs significativos (ex: status, cost_price).
type LegacyProduct struct {
	ID               int64
	SKU              string
	Name             string
	ShortDescription sql.NullString
	FullDescription  sql.NullString
	Status           sql.NullBool
	UOM              sql.NullString
	CostPrice        sql.NullFloat64
	CurrentStock     sql.NullInt64
	MinimumStock     sql.NullInt64
	Weight           sql.NullFloat64
	HeightCm         sql.NullFloat64
	WidthCm          sql.NullFloat64
	LengthCm         sql.NullFloat64
	Height           sql.NullFloat64
	Width            sql.NullFloat64
	Length           sql.NullFloat64
	EAN              sql.NullString
	NCM              sql.NullString
	IsVariant        bool
	ParentID         sql.NullInt64
	TenantID         int64
}

// ReadAll lê todas as linhas da tabela products do bancoProd.db, em ordem de id.
func ReadAll(db *sql.DB) ([]LegacyProduct, error) {
	rows, err := db.Query(`
		SELECT
			id, sku, name, short_description, full_description, status, uom,
			cost_price, current_stock, minimum_stock,
			weight, height_cm, width_cm, length_cm, height, width, length,
			ean, ncm, is_variant, parent_id, tenant_id
		FROM products
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LegacyProduct
	for rows.Next() {
		var p LegacyProduct
		if err := rows.Scan(
			&p.ID, &p.SKU, &p.Name, &p.ShortDescription, &p.FullDescription, &p.Status, &p.UOM,
			&p.CostPrice, &p.CurrentStock, &p.MinimumStock,
			&p.Weight, &p.HeightCm, &p.WidthCm, &p.LengthCm, &p.Height, &p.Width, &p.Length,
			&p.EAN, &p.NCM, &p.IsVariant, &p.ParentID, &p.TenantID,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// LegacyVariant é uma linha da tabela `variants` do bancoProd.db — schema
// real de variação por produto (cor/tamanho/etc), separado da tabela
// `products` (que tem colunas cost_price/current_stock/etc próprias, mas
// mortas/zeradas pra produtos que já têm linha em `variants`).
type LegacyVariant struct {
	ID           int64
	SKU          string
	CostPrice    sql.NullFloat64
	CurrentStock sql.NullInt64
	MinimumStock sql.NullInt64
	Weight       sql.NullFloat64
	HeightCm     sql.NullFloat64
	WidthCm      sql.NullFloat64
	LengthCm     sql.NullFloat64
	IsDefault    bool
	ProductID    int64
	TenantID     int64
	Status       sql.NullBool
	EAN          sql.NullString
}

// ReadVariants lê as variantes ativas (deleted_at IS NULL) da tabela `variants`.
func ReadVariants(db *sql.DB) ([]LegacyVariant, error) {
	rows, err := db.Query(`
		SELECT
			id, sku, cost_price, current_stock, minimum_stock,
			weight, height_cm, width_cm, length_cm, is_default,
			product_id, tenant_id, status, ean
		FROM variants
		WHERE deleted_at IS NULL
		ORDER BY product_id, id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LegacyVariant
	for rows.Next() {
		var v LegacyVariant
		if err := rows.Scan(
			&v.ID, &v.SKU, &v.CostPrice, &v.CurrentStock, &v.MinimumStock,
			&v.Weight, &v.HeightCm, &v.WidthCm, &v.LengthCm, &v.IsDefault,
			&v.ProductID, &v.TenantID, &v.Status, &v.EAN,
		); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// LegacyAttribute é uma linha de `attributes` (ex: "Cor", "Tamanho").
type LegacyAttribute struct {
	ID       int64
	Name     string
	TenantID int64
}

func ReadAttributes(db *sql.DB) ([]LegacyAttribute, error) {
	rows, err := db.Query(`SELECT id, name, tenant_id FROM attributes ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LegacyAttribute
	for rows.Next() {
		var a LegacyAttribute
		if err := rows.Scan(&a.ID, &a.Name, &a.TenantID); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// LegacyAttributeValue é uma linha de `attribute_values` (ex: "Vermelho", "P").
type LegacyAttributeValue struct {
	ID          int64
	Value       string
	AttributeID int64
}

func ReadAttributeValues(db *sql.DB) ([]LegacyAttributeValue, error) {
	rows, err := db.Query(`SELECT id, value, attribute_id FROM attribute_values ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LegacyAttributeValue
	for rows.Next() {
		var v LegacyAttributeValue
		if err := rows.Scan(&v.ID, &v.Value, &v.AttributeID); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// LegacyVariantAttribute é uma linha do join `variant_attributes`.
type LegacyVariantAttribute struct {
	VariantID        int64
	AttributeValueID int64
}

func ReadVariantAttributes(db *sql.DB) ([]LegacyVariantAttribute, error) {
	rows, err := db.Query(`SELECT variant_id, attribute_value_id FROM variant_attributes ORDER BY variant_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LegacyVariantAttribute
	for rows.Next() {
		var va LegacyVariantAttribute
		if err := rows.Scan(&va.VariantID, &va.AttributeValueID); err != nil {
			return nil, err
		}
		out = append(out, va)
	}
	return out, rows.Err()
}
