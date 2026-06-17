package services

import "github.com/DTineli/ez/internal/store"

// ApplyPriceTable aplica o multiplicador da tabela ao custo base.
// Retorna costPrice sem alteração se pt for nil.
func ApplyPriceTable(costPrice float64, pt *store.PriceTable) float64 {
	if pt == nil {
		return costPrice
	}
	return costPrice * (1 + pt.Percentage/100)
}
