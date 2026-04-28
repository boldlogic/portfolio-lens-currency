package currencies

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/JohannesJHN/iso4217"
)

var (
	ErrWrongISOCharCode    = errors.New("символьный код валюты по ISO 4217 состоит из 3 латинских букв")
	ErrNotExistingCurrency = errors.New("валюта не существует в ISO 4217")
)

// NewCurrencyCode нормализует и проверяет символьный код валюты по ISO 4217.
func NewCurrencyCode(charCode string) (CurrencyCode, error) {
	code := normalizeCurrencyCode(charCode)
	if err := validateNormalizedCurrencyCode(code); err != nil {
		return "", err
	}
	return CurrencyCode(code), nil
}

// CheckCurrencyCode проверяет, что код валюты корректен и существует в ISO 4217.
func CheckCurrencyCode(charCode string) error {
	_, err := NewCurrencyCode(charCode)
	return err
}

// String возвращает строковое представление кода валюты.
func (c CurrencyCode) String() string {
	return string(c)
}

// Validate проверяет корректность значения CurrencyCode.
func (c CurrencyCode) Validate() error {
	_, err := NewCurrencyCode(c.String())
	return err
}

// NewCurrencyPair создает валютную пару и проверяет оба кода валют.
func NewCurrencyPair(base, quote string) (CurrencyPair, error) {
	baseCode, err := NewCurrencyCode(base)
	if err != nil {
		return CurrencyPair{}, fmt.Errorf("проверить базовую валюту: %w", err)
	}

	quoteCode, err := NewCurrencyCode(quote)
	if err != nil {
		return CurrencyPair{}, fmt.Errorf("проверить валюту котировки: %w", err)
	}

	return CurrencyPair{
		Base:  baseCode,
		Quote: quoteCode,
	}, nil
}

// Validate проверяет корректность базовой валюты и валюты котировки в паре.
func (p CurrencyPair) Validate() error {
	if err := p.Base.Validate(); err != nil {
		return fmt.Errorf("проверить базовую валюту: %w", err)
	}
	if err := p.Quote.Validate(); err != nil {
		return fmt.Errorf("проверить валюту котировки: %w", err)
	}
	return nil
}

// normalizeCurrencyCode приводит код валюты к верхнему регистру и убирает пробелы по краям.
func normalizeCurrencyCode(charCode string) string {
	return strings.ToUpper(strings.TrimSpace(charCode))
}

// validateNormalizedCurrencyCode проверяет длину, алфавит и наличие кода в ISO 4217.
func validateNormalizedCurrencyCode(code string) error {
	if utf8.RuneCountInString(code) != 3 {
		return ErrWrongISOCharCode
	}

	for _, r := range code {
		if r < 'A' || r > 'Z' {
			return ErrWrongISOCharCode
		}
	}
	if _, ok := iso4217.LookupByAlpha3(code); !ok {
		return fmt.Errorf("%w: код=%s", ErrNotExistingCurrency, code)
	}
	return nil
}
