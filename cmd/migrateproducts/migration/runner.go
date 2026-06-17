package migration

import (
	"fmt"
	"strings"

	"github.com/DTineli/ez/cmd/migrateproducts/legacy"
	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/store/dbstore"
	"gorm.io/gorm"
)

// Summary é o relatório final impresso pro operador.
type Summary struct {
	RowsRead         int
	ProductsCreated  int
	VariantsCreated  int
	VariantsUpdated  int
	AttributesLinked int
	Skipped          int
	Warnings         []string
	SkippedDetails   []string
}

// TenantGetter é o subconjunto de store.TenantStore que essa migração usa.
type TenantGetter interface {
	GetTenantByID(id uint) (*store.Tenant, error)
}

// ValidateTenants confere que todo tenant_id presente nos produtos mapeados
// já existe no banco de destino. Retorna a lista de tenants faltando — se
// não for vazia, a migração deve abortar sem escrever nada.
func ValidateTenants(mapped []MappedProduct, tenants TenantGetter) (missing []uint, err error) {
	seen := make(map[uint]bool)
	for _, m := range mapped {
		id := m.Product.TenantID
		if seen[id] {
			continue
		}
		seen[id] = true

		if _, getErr := tenants.GetTenantByID(id); getErr != nil {
			missing = append(missing, id)
		}
	}
	return missing, nil
}

// isDuplicateError detecta violação de unique constraint entre os drivers
// usados no repo (mesma lógica de internal/handlers/helpers.go).
func isDuplicateError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") ||
		strings.Contains(msg, "Duplicate") ||
		strings.Contains(msg, "duplicate key value violates unique constraint")
}

// existingProductID busca, só leitura, se já existe Product com esse
// tenant_id+sku em prod (não tem método pronto no ProductStore pra isso).
func existingProductID(db *gorm.DB, tenantID uint, sku string) (uint, bool) {
	var p store.Product
	err := db.Select("id").Where("tenant_id = ? AND sku = ?", tenantID, sku).First(&p).Error
	if err != nil {
		return 0, false
	}
	return p.ID, true
}

// existingVariantID busca, só leitura, se já existe Variant com esse
// tenant_id+sku (não-deletado) em prod.
func existingVariantID(db *gorm.DB, tenantID uint, sku string) (uint, bool) {
	var v store.Variant
	err := db.Select("id").Where("tenant_id = ? AND sku = ?", tenantID, sku).First(&v).Error
	if err != nil {
		return 0, false
	}
	return v.ID, true
}

// Run insere/corrige os produtos mapeados em prod. Idempotente:
//   - produto novo (tenant+sku não existe ainda): CreateProduct cria
//     Product+Variants em cascata.
//   - produto já existente (caso do bancoProd já ter sido rodado uma vez
//     com o mapper antigo, que criava 1 variant default zerada por
//     produto): reaproveita o Product, e por variant:
//   - se já existe variant com esse sku → UpdateVariantFields com os
//     dados corretos (corrige a variant fake criada antes).
//   - senão → CreateVariant (variants reais adicionais, ex: cor/tamanho).
//
// Em dry-run, só simula e não escreve nada.
func Run(db *gorm.DB, products *dbstore.ProductStore, mapped []MappedProduct, dryRun bool) (Summary, map[int64]uint) {
	s := Summary{RowsRead: len(mapped)}
	oldVariantIDToNew := make(map[int64]uint) // pra linkar variant_attributes depois

	for _, m := range mapped {
		// Lookup é só leitura — roda mesmo em dry-run pra relatório ficar
		// fiel ao que já existe em prod (importante aqui já que rodaram
		// uma vez com o mapper antigo e criaram dados parciais/errados).
		productID, productExists := existingProductID(db, m.Product.TenantID, m.Product.SKU)

		if !productExists {
			product := m.Product
			for _, mv := range m.Variants {
				product.Variants = append(product.Variants, mv.Variant)
			}

			if dryRun {
				s.ProductsCreated++
				s.VariantsCreated += len(product.Variants)
				for _, mv := range m.Variants {
					if mv.OldID != 0 {
						oldVariantIDToNew[mv.OldID] = 0 // placeholder, só pra contar links no dry-run
					}
				}
				continue
			}

			if err := products.CreateProduct(&product); err != nil {
				if isDuplicateError(err) {
					s.Skipped++
					s.SkippedDetails = append(s.SkippedDetails, fmt.Sprintf(
						"old product id=%d sku=%s tenant=%d: já existe, pulado",
						m.OldID, product.SKU, product.TenantID,
					))
					continue
				}
				s.Skipped++
				s.SkippedDetails = append(s.SkippedDetails, fmt.Sprintf(
					"old product id=%d sku=%s tenant=%d: erro ao criar: %v",
					m.OldID, product.SKU, product.TenantID, err,
				))
				continue
			}

			s.ProductsCreated++
			s.VariantsCreated += len(product.Variants)
			for i, mv := range m.Variants {
				if mv.OldID != 0 {
					oldVariantIDToNew[mv.OldID] = product.Variants[i].ID
				}
			}
			continue
		}

		// Produto já existe em prod — corrige/completa as variants.
		for _, mv := range m.Variants {
			variant := mv.Variant

			if dryRun {
				if _, found := existingVariantID(db, variant.TenantID, variant.SKU); found {
					s.VariantsUpdated++
				} else {
					s.VariantsCreated++
				}
				if mv.OldID != 0 {
					oldVariantIDToNew[mv.OldID] = 0 // placeholder, só pra contar links no dry-run
				}
				continue
			}

			if existingID, found := existingVariantID(db, variant.TenantID, variant.SKU); found {
				fields := map[string]any{
					"cost_price":    variant.CostPrice,
					"current_stock": variant.CurrentStock,
					"weight":        variant.Weight,
					"height_cm":     variant.HeightCm,
					"width_cm":      variant.WidthCm,
					"length_cm":     variant.LengthCm,
					"ean":           variant.EAN,
					"status":        variant.Status,
					"is_default":    variant.IsDefault,
				}
				if err := products.UpdateVariantFields(existingID, variant.TenantID, fields); err != nil {
					s.Skipped++
					s.SkippedDetails = append(s.SkippedDetails, fmt.Sprintf(
						"old variant id=%d sku=%s: erro ao corrigir: %v",
						mv.OldID, variant.SKU, err,
					))
					continue
				}
				s.VariantsUpdated++
				if mv.OldID != 0 {
					oldVariantIDToNew[mv.OldID] = existingID
				}
				continue
			}

			variant.ProductID = productID
			if err := products.CreateVariant(&variant); err != nil {
				if isDuplicateError(err) {
					s.Skipped++
					s.SkippedDetails = append(s.SkippedDetails, fmt.Sprintf(
						"old variant id=%d sku=%s: já existe, pulado",
						mv.OldID, variant.SKU,
					))
					continue
				}
				s.Skipped++
				s.SkippedDetails = append(s.SkippedDetails, fmt.Sprintf(
					"old variant id=%d sku=%s: erro ao criar: %v",
					mv.OldID, variant.SKU, err,
				))
				continue
			}
			s.VariantsCreated++
			if mv.OldID != 0 {
				oldVariantIDToNew[mv.OldID] = variant.ID
			}
		}
	}

	return s, oldVariantIDToNew
}

// LinkAttributes recria attributes/attribute_values/variant_attributes em
// prod a partir do legacy, usando FindOrCreateAttribute/FindOrCreateAttributeValue
// (idempotentes — usam FirstOrCreate) e SetVariantAttributes (idempotente —
// substitui o conjunto de attribute_values da variant a cada chamada).
// oldVariantIDToNew vem do retorno de Run.
func LinkAttributes(
	products *dbstore.ProductStore,
	attrs []legacy.LegacyAttribute,
	values []legacy.LegacyAttributeValue,
	links []legacy.LegacyVariantAttribute,
	oldVariantIDToNew map[int64]uint,
	dryRun bool,
) (linked int, warnings []string) {
	oldAttrIDToNew := make(map[int64]uint)
	for _, a := range attrs {
		if dryRun {
			oldAttrIDToNew[a.ID] = uint(a.ID) // placeholder, só pra contagem
			continue
		}
		newAttr, err := products.FindOrCreateAttribute(a.Name, uint(a.TenantID))
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("attribute %q (old id=%d): erro: %v", a.Name, a.ID, err))
			continue
		}
		oldAttrIDToNew[a.ID] = newAttr.ID
	}

	oldValueIDToNew := make(map[int64]uint)
	for _, v := range values {
		newAttrID, ok := oldAttrIDToNew[v.AttributeID]
		if !ok {
			warnings = append(warnings, fmt.Sprintf("attribute_value %q (old id=%d): attribute pai não mapeado", v.Value, v.ID))
			continue
		}
		if dryRun {
			oldValueIDToNew[v.ID] = uint(v.ID) // placeholder
			continue
		}
		newVal, err := products.FindOrCreateAttributeValue(v.Value, newAttrID)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("attribute_value %q (old id=%d): erro: %v", v.Value, v.ID, err))
			continue
		}
		oldValueIDToNew[v.ID] = newVal.ID
	}

	valueIDsByVariant := make(map[int64][]uint)
	for _, l := range links {
		newValueID, ok := oldValueIDToNew[l.AttributeValueID]
		if !ok {
			continue
		}
		valueIDsByVariant[l.VariantID] = append(valueIDsByVariant[l.VariantID], newValueID)
	}

	for oldVariantID, valueIDs := range valueIDsByVariant {
		newVariantID, ok := oldVariantIDToNew[oldVariantID]
		if !ok {
			warnings = append(warnings, fmt.Sprintf(
				"variant_attributes: old variant id=%d não foi mapeado (variant não migrada?), pulando %d link(s)",
				oldVariantID, len(valueIDs),
			))
			continue
		}
		if dryRun {
			linked += len(valueIDs)
			continue
		}
		if err := products.SetVariantAttributes(newVariantID, valueIDs); err != nil {
			warnings = append(warnings, fmt.Sprintf("variant id=%d: erro ao linkar atributos: %v", newVariantID, err))
			continue
		}
		linked += len(valueIDs)
	}

	return linked, warnings
}
