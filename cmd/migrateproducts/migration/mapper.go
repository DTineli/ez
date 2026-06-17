// Package migration mapeia o schema legacy (bancoProd.db) pra
// store.Product/store.Variant/store.Attribute do schema atual e roda a
// inserção em prod.
package migration

import (
	"database/sql"
	"fmt"

	"github.com/DTineli/ez/cmd/migrateproducts/legacy"
	"github.com/DTineli/ez/internal/store"
)

// IgnoredTenantID é o tenant que sabidamente não deve ser migrado
// (decisão do usuário: linha única, considerada lixo de teste).
const IgnoredTenantID = 5

// MappedVariant carrega o OldID (ID na tabela `variants` do bancoProd) junto
// com o Variant mapeado, pra depois linkar variant_attributes corretamente.
// OldID == 0 significa variant sintetizada a partir das colunas flat de
// `products` (produto sem nenhuma linha em `variants` no legacy) — nesse
// caso não existe variant_attributes pra linkar.
type MappedVariant struct {
	OldID   int64
	Variant store.Variant
}

// MappedProduct é o resultado de mapear um produto + suas variantes.
type MappedProduct struct {
	OldID    int64
	Product  store.Product
	Variants []MappedVariant
}

// MapAll mapeia products+variants legacy pro schema atual. Pra cada produto:
//   - se existir linha(s) em `variants` (real) pra ele, usa elas direto —
//     são a fonte de verdade de cost_price/current_stock/dimensões/ean/status.
//   - se não existir nenhuma, sintetiza 1 variant default a partir das
//     colunas flat da própria `products` (legado de produtos sem variação).
//
// Linhas do tenant ignorado são puladas e reportadas em warnings.
func MapAll(products []legacy.LegacyProduct, variants []legacy.LegacyVariant) (mapped []MappedProduct, warnings []string) {
	variantsByProduct := make(map[int64][]legacy.LegacyVariant)
	for _, v := range variants {
		variantsByProduct[v.ProductID] = append(variantsByProduct[v.ProductID], v)
	}

	for _, p := range products {
		if p.TenantID == IgnoredTenantID {
			warnings = append(warnings, fmt.Sprintf(
				"row id=%d (sku=%s): tenant_id=%d ignorado por decisão do usuário, pulando",
				p.ID, p.SKU, IgnoredTenantID,
			))
			continue
		}

		status, sw := mapProductStatus(p)
		warnings = append(warnings, sw...)

		uom := store.Unit
		if p.UOM.Valid && p.UOM.String != "" {
			uom = store.UOM(p.UOM.String)
		}

		product := store.Product{
			SKU:             p.SKU,
			Name:            p.Name,
			FullDescription: p.FullDescription.String,
			Status:          status,
			UOM:             uom,
			NCM:             p.NCM.String,
			// Product.Height/Width/Length ficam 0 — dimensões reais vão pra Variant
			// (decisão do usuário, ver plano de migração).
			TenantID: uint(p.TenantID),
		}

		realVariants := variantsByProduct[p.ID]

		var mappedVariants []MappedVariant
		if len(realVariants) == 0 {
			// Produto sem linha em `variants`: sintetiza 1 default a partir
			// das colunas flat de `products` (era o único dado disponível).
			variant, vw := synthesizeVariantFromProduct(p, status)
			warnings = append(warnings, vw...)
			mappedVariants = append(mappedVariants, MappedVariant{OldID: 0, Variant: variant})
		} else {
			// Decisão do usuário: default só existe pra produto criado sem
			// variação (caso sintetizado abaixo). Produto com variação real
			// não precisa de uma marcada default — migra is_default como
			// veio do legacy, sem forçar nenhuma.
			for _, v := range realVariants {
				variant, vw := mapRealVariant(v)
				warnings = append(warnings, vw...)
				mappedVariants = append(mappedVariants, MappedVariant{OldID: v.ID, Variant: variant})
			}
		}

		mapped = append(mapped, MappedProduct{OldID: p.ID, Product: product, Variants: mappedVariants})
	}

	return mapped, warnings
}

// synthesizeVariantFromProduct constrói uma variant default a partir das
// colunas flat de `products`, pra produtos legacy que nunca tiveram linha
// em `variants` (sem variação).
func synthesizeVariantFromProduct(p legacy.LegacyProduct, status bool) (store.Variant, []string) {
	var warnings []string

	if p.MinimumStock.Valid && p.MinimumStock.Int64 != 0 {
		warnings = append(warnings, fmt.Sprintf(
			"row id=%d (sku=%s): minimum_stock=%d descartado (campo não existe no schema atual)",
			p.ID, p.SKU, p.MinimumStock.Int64,
		))
	}
	if p.ShortDescription.Valid && p.ShortDescription.String != "" {
		warnings = append(warnings, fmt.Sprintf(
			"row id=%d (sku=%s): short_description descartado (campo não existe no schema atual)",
			p.ID, p.SKU,
		))
	}

	return store.Variant{
		SKU:          p.SKU,
		Status:       status,
		CostPrice:    p.CostPrice.Float64,
		CurrentStock: int(p.CurrentStock.Int64),
		EAN:          p.EAN.String,
		Weight:       p.Weight.Float64,
		HeightCm:     fallback(p.Height, p.HeightCm),
		WidthCm:      fallback(p.Width, p.WidthCm),
		LengthCm:     fallback(p.Length, p.LengthCm),
		IsDefault:    true,
		TenantID:     uint(p.TenantID),
	}, warnings
}

// mapRealVariant mapeia uma linha real de `variants` (1:1, sem necessidade
// de fallback — os campos *_cm já vêm populados nessa tabela).
func mapRealVariant(v legacy.LegacyVariant) (store.Variant, []string) {
	var warnings []string

	status := true // Variant.Status default:true no schema atual
	if v.Status.Valid {
		status = v.Status.Bool
	} else {
		warnings = append(warnings, fmt.Sprintf(
			"variant old id=%d (sku=%s): status NULL, default para true (ativo)",
			v.ID, v.SKU,
		))
	}

	if v.MinimumStock.Valid && v.MinimumStock.Int64 != 0 {
		warnings = append(warnings, fmt.Sprintf(
			"variant old id=%d (sku=%s): minimum_stock=%d descartado (campo não existe no schema atual)",
			v.ID, v.SKU, v.MinimumStock.Int64,
		))
	}

	return store.Variant{
		SKU:          v.SKU,
		Status:       status,
		CostPrice:    v.CostPrice.Float64,
		CurrentStock: int(v.CurrentStock.Int64),
		EAN:          v.EAN.String,
		Weight:       v.Weight.Float64,
		HeightCm:     v.HeightCm.Float64,
		WidthCm:      v.WidthCm.Float64,
		LengthCm:     v.LengthCm.Float64,
		IsDefault:    v.IsDefault,
		TenantID:     uint(v.TenantID),
	}, warnings
}

func mapProductStatus(p legacy.LegacyProduct) (bool, []string) {
	if !p.Status.Valid {
		return false, []string{fmt.Sprintf(
			"row id=%d (sku=%s): status NULL, default para false (inativo)",
			p.ID, p.SKU,
		)}
	}
	return p.Status.Bool, nil
}

// fallback usa `primary` (colunas height/width/length sem sufixo da tabela
// `products`, que têm os dados reais nesse dataset pros produtos sem
// variant) e cai pra `cm` (quase sempre vazias) só se primary for NULL/zero.
func fallback(primary, cm sql.NullFloat64) float64 {
	if primary.Valid && primary.Float64 != 0 {
		return primary.Float64
	}
	return cm.Float64
}
