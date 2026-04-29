package templates

import "github.com/DTineli/ez/internal/store"

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
