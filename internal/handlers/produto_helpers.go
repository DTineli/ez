package handlers

import (
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/DTineli/ez/internal/forms"
	"github.com/DTineli/ez/internal/store"
)

func validateProductForm(r *http.Request) (*forms.Form, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	form := forms.New(r.PostForm)

	form.Required("name", "sku")
	form.MaxLength("name", 255)
	form.MinLength("name", 4)

	form.MaxLength("uom", 2)
	form.MinLength("uom", 2)

	form.MaxLength("description", 15000)

	form.MaxLength("sku", 25)
	form.MinLength("sku", 4)

	form.IsInt("ean")
	form.Set("ncm", strings.ReplaceAll(form.Get("ncm"), ".", ""))
	form.IsInt("ncm")

	return form, nil
}

type parsedVariation struct {
	name    string
	options []parsedOption
}

type parsedOption struct {
	value       string
	description string
}

type comboInput struct {
	price float64
	stock int
	sku   string
	gtin  string
}

func parseVariations(form url.Values) []parsedVariation {
	varNameRe := regexp.MustCompile(`^variation\[(\d+)\]\[name\]$`)
	optValueRe := regexp.MustCompile(`^variation\[(\d+)\]\[option\]\[(\d+)\]\[value\]$`)
	optDescRe := regexp.MustCompile(`^variation\[(\d+)\]\[option\]\[(\d+)\]\[description\]$`)

	varNames := make(map[int]string)
	optValues := make(map[int]map[int]string)
	optDescs := make(map[int]map[int]string)
	var maxVarIdx int

	for key, vals := range form {
		if len(vals) == 0 {
			continue
		}
		v := vals[0]
		if m := varNameRe.FindStringSubmatch(key); m != nil {
			idx, _ := strconv.Atoi(m[1])
			varNames[idx] = v
			if idx > maxVarIdx {
				maxVarIdx = idx
			}
		} else if m := optValueRe.FindStringSubmatch(key); m != nil {
			vi, _ := strconv.Atoi(m[1])
			oi, _ := strconv.Atoi(m[2])
			if optValues[vi] == nil {
				optValues[vi] = make(map[int]string)
			}
			optValues[vi][oi] = v
		} else if m := optDescRe.FindStringSubmatch(key); m != nil {
			vi, _ := strconv.Atoi(m[1])
			oi, _ := strconv.Atoi(m[2])
			if optDescs[vi] == nil {
				optDescs[vi] = make(map[int]string)
			}
			optDescs[vi][oi] = v
		}
	}

	result := make([]parsedVariation, 0)
	for vi := 0; vi <= maxVarIdx; vi++ {
		name := varNames[vi]
		if name == "" {
			continue
		}
		var maxOptIdx int
		for oi := range optValues[vi] {
			if oi > maxOptIdx {
				maxOptIdx = oi
			}
		}
		opts := make([]parsedOption, 0)
		for oi := 0; oi <= maxOptIdx; oi++ {
			val := optValues[vi][oi]
			if val == "" {
				continue
			}
			opts = append(opts, parsedOption{value: val, description: optDescs[vi][oi]})
		}
		if len(opts) == 0 {
			continue
		}
		result = append(result, parsedVariation{name: name, options: opts})
	}
	return result
}

func parseComboInputs(form url.Values, count int) []comboInput {
	result := make([]comboInput, count)
	comboRe := regexp.MustCompile(`^combo\[(\d+)\]\[(\w+)\]$`)
	for key, vals := range form {
		if len(vals) == 0 {
			continue
		}
		m := comboRe.FindStringSubmatch(key)
		if m == nil {
			continue
		}
		ci, _ := strconv.Atoi(m[1])
		if ci >= count {
			continue
		}
		switch m[2] {
		case "price":
			result[ci].price, _ = strconv.ParseFloat(vals[0], 64)
		case "stock":
			result[ci].stock, _ = strconv.Atoi(vals[0])
		case "sku":
			result[ci].sku = vals[0]
		case "gtin":
			result[ci].gtin = vals[0]
		}
	}
	return result
}

func cartesianProductIDs(sets [][]uint) [][]uint {
	if len(sets) == 0 {
		return nil
	}
	result := [][]uint{{}}
	for _, set := range sets {
		if len(set) == 0 {
			continue
		}
		newResult := make([][]uint, 0, len(result)*len(set))
		for _, existing := range result {
			for _, item := range set {
				combo := make([]uint, len(existing)+1)
				copy(combo, existing)
				combo[len(existing)] = item
				newResult = append(newResult, combo)
			}
		}
		result = newResult
	}
	return result
}

func variantHasExactAttributes(v store.Variant, ids []uint) bool {
	if len(v.Attributes) != len(ids) {
		return false
	}
	existing := make(map[uint]bool, len(v.Attributes))
	for _, va := range v.Attributes {
		existing[va.AttributeValueID] = true
	}
	for _, id := range ids {
		if !existing[id] {
			return false
		}
	}
	return true
}

func mapProductToForm(p *store.Product) *forms.Form {
	form := forms.New(nil)

	form.Set("ID", strconv.Itoa(int(p.ID)))
	form.Set("name", p.Name)
	form.Set("sku", p.SKU)
	form.Set("uom", string(p.UOM))
	form.Set("description", p.FullDescription)
	form.Set("ean", p.EAN)

	ncm := p.NCM
	if len(ncm) == 8 {
		ncm = ncm[:4] + "." + ncm[4:6] + "." + ncm[6:]
	}
	form.Set("ncm", ncm)

	return form
}
