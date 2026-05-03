package templates

import (
	"encoding/json"
	"strings"

	"github.com/DTineli/ez/internal/store"
)

type variantJS struct {
	ID    uint    `json:"id"`
	Price float64 `json:"price"`
	Label string  `json:"label"`
}

func VariantLabel(v store.VariantData) string {
	if len(v.Attrs) == 0 {
		return "Padrão"
	}
	parts := make([]string, 0, len(v.Attrs))
	for _, a := range v.Attrs {
		parts = append(parts, a.Value)
	}
	return strings.Join(parts, " / ")
}

func BuildVariantsJS(variants []store.VariantData) string {
	if len(variants) == 0 {
		return `[{"id":0,"price":0,"label":"Padrão"}]`
	}
	items := make([]variantJS, 0, len(variants))
	for _, v := range variants {
		items = append(items, variantJS{
			ID:    v.ID,
			Price: v.Price,
			Label: VariantLabel(v),
		})
	}
	b, _ := json.Marshal(items)
	return string(b)
}

type VariationCardData struct {
	Name    string                `json:"name"`
	Options []VariationOptionData `json:"options"`
}

type VariationOptionData struct {
	Value string `json:"value"`
}

func BuildVariationCards(variants []store.Variant) []VariationCardData {
	type entry struct {
		name     string
		values   []string
		valueSet map[string]bool
	}

	attrMap := make(map[uint]*entry)
	attrOrder := []uint{}

	for _, v := range variants {
		for _, va := range v.Attributes {
			attr := va.AttributeValue.Attribute
			if _, exists := attrMap[attr.ID]; !exists {
				attrOrder = append(attrOrder, attr.ID)
				attrMap[attr.ID] = &entry{
					name:     attr.Name,
					valueSet: make(map[string]bool),
				}
			}
			e := attrMap[attr.ID]
			val := va.AttributeValue.Value
			if !e.valueSet[val] {
				e.valueSet[val] = true
				e.values = append(e.values, val)
			}
		}
	}

	result := make([]VariationCardData, 0, len(attrOrder))
	for _, id := range attrOrder {
		e := attrMap[id]
		opts := make([]VariationOptionData, 0, len(e.values))
		for _, v := range e.values {
			opts = append(opts, VariationOptionData{Value: v})
		}
		result = append(result, VariationCardData{Name: e.name, Options: opts})
	}

	return result
}
