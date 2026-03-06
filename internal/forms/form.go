package forms

import (
	"fmt"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Form struct {
	url.Values
	Errors errors
}

func New(data url.Values) *Form {
	if data == nil {
		data = url.Values{}
	}

	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		if strings.TrimSpace(f.Get(field)) == "" {
			f.Errors[field] = append(f.Errors[field], "Esse valor não pode ser vazio.")
		}
	}
}

func (f *Form) IsEmail(field string) {
	email := strings.TrimSpace(f.Get(field))

	if email == "" {
		return
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		f.Errors[field] = append(f.Errors[field], fmt.Sprintf("%v deve ser um email valido", field))
	}
}

func (f *Form) MaxLength(field string, length int) {
	if utf8.RuneCountInString(f.Get(field)) > length {
		f.Errors[field] = append(f.Errors[field], fmt.Sprintf("%v deve ter no maximo %v caracteres", field, length))
	}
}

func (f *Form) MinLength(field string, length int) {
	value := f.Get(field)
	if value == "" {
		return
	}

	if utf8.RuneCountInString(value) < length {
		f.Errors[field] = append(f.Errors[field], fmt.Sprintf("%v deve ter no minimo %v caracteres", field, length))
	}
}

func (f *Form) IsFloat(field string) float64 {
	value := f.Get(field)

	if value == "" {
		return 0
	}

	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		f.Errors[field] = append(f.Errors[field], "Deve ser um número válido")
		return 0
	}

	return val
}

func (f *Form) IsInt(field string) int {
	value := f.Get(field)

	if value == "" {
		return 0
	}

	val, err := strconv.Atoi(value)
	if err != nil {
		f.Errors[field] = append(f.Errors[field], "Deve ser um número inteiro válido")
	}

	return val
}

func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}
