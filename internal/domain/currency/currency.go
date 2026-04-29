package currency

import (
	"errors"

	"github.com/boldlogic/portfolio-lens-currency/pkg/currencies"
)

var ErrInvalidCurrency = errors.New("некорректная валюта")

type Currency struct {
	Code        currencies.CurrencyCode
	NumericCode int16
	Name        *string
	LatinName   string
	MinorUnits  int16
	Active      bool
}

type NewCurrencyInput struct {
	Code        string
	NumericCode int16
	Name        *string
	LatinName   string
	MinorUnits  int16
	Active      bool
}

func NewCurrency(input NewCurrencyInput) (Currency, error) {
	code, err := currencies.ParseCurrencyCode(input.Code)
	if err != nil {
		return Currency{}, err
	}
	if input.NumericCode <= 0 || input.MinorUnits < 0 {
		return Currency{}, ErrInvalidCurrency
	}

	return Currency{
		Code:        code,
		NumericCode: input.NumericCode,
		Name:        input.Name,
		LatinName:   input.LatinName,
		MinorUnits:  input.MinorUnits,
		Active:      input.Active,
	}, nil
}
