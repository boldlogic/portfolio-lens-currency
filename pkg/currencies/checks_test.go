package currencies

import (
	"errors"
	"testing"
)

func TestParseCurrencyCode(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want CurrencyCode
	}{
		{"верхний_регистр", "USD", "USD"},
		{"нижний_регистр", "eur", "EUR"},
		{"смешанный_регистр", "Usd", "USD"},
		{"пробелы_по_краям", " rub ", "RUB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCurrencyCode(tt.in)
			if err != nil {
				t.Fatalf("ParseCurrencyCode(%q) вернула ошибку: %v", tt.in, err)
			}
			if got != tt.want {
				t.Fatalf("ParseCurrencyCode(%q) = %q, ожидали %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestCheckCurrencyCode(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"верхний_регистр", "USD"},
		{"нижний_регистр", "eur"},
		{"смешанный_регистр", "Usd"},
		{"пробелы_по_краям", " rub "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CheckCurrencyCode(tt.code); err != nil {
				t.Fatalf("CheckCurrencyCode(%q) вернула ошибку: %v", tt.code, err)
			}
		})
	}
}

func TestCheckCurrencyCodeWrongISOCharCode(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"две_буквы", "US"},
		{"четыре_буквы", "USDD"},
		{"цифры", "US1"},
		{"кириллица", "РУБ"},
		{"пусто", ""},
		{"пробел", "  "},
		{"спецсимвол", "U$D"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckCurrencyCode(tt.code)
			if !errors.Is(err, ErrWrongISOCharCode) {
				t.Fatalf("CheckCurrencyCode(%q) = %v, ожидали %v", tt.code, err, ErrWrongISOCharCode)
			}
		})
	}
}

func TestCheckCurrencyCodeErrNotExistingCurrency(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"AAA", "AAA"},
		{"aaa", "aaa"},
		{"GLD", "GLD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckCurrencyCode(tt.code)
			if !errors.Is(err, ErrNotExistingCurrency) {
				t.Fatalf("CheckCurrencyCode(%q) = %v, ожидали %v", tt.code, err, ErrNotExistingCurrency)
			}
		})
	}
}

func TestNewCurrencyPair(t *testing.T) {
	pair, err := NewCurrencyPair("usd", "rub")
	if err != nil {
		t.Fatalf("NewCurrencyPair вернула ошибку: %v", err)
	}
	if pair.Base != "USD" || pair.Quote != "RUB" {
		t.Fatalf("NewCurrencyPair = %+v, ожидали USD/RUB", pair)
	}
}
