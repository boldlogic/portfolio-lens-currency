package service

import (
	"context"
	"testing"
	"time"

	"github.com/boldlogic/portfolio-lens-currency/internal/domain/currency"
	"github.com/boldlogic/portfolio-lens-currency/internal/domain/rate"
	"github.com/boldlogic/portfolio-lens-currency/pkg/currencies"
)

type currencyRepositoryStub struct{}

func (currencyRepositoryStub) GetCurrency(context.Context, currencies.CurrencyCode) (currency.Currency, error) {
	return currency.Currency{}, nil
}

type rateRepositoryStub struct {
	rate rate.Rate
}

func (s *rateRepositoryStub) GetRate(context.Context, rate.Query) (rate.Rate, error) {
	return s.rate, nil
}

func (s *rateRepositoryStub) SaveRate(context.Context, rate.Rate) error {
	return nil
}

func TestServiceConvert(t *testing.T) {
	fxRate, err := rate.NewRate(rate.NewRateInput{
		SourceCode: "cbr",
		Base:       "USD",
		Quote:      "RUB",
		Value:      "90.5",
		ObservedAt: time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC),
		ValidFrom:  time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("rate.NewRate вернула ошибку: %v", err)
	}

	service, err := New(currencyRepositoryStub{}, &rateRepositoryStub{rate: fxRate})
	if err != nil {
		t.Fatalf("New вернула ошибку: %v", err)
	}

	result, err := service.Convert(context.Background(), "2", "usd", "rub", nil)
	if err != nil {
		t.Fatalf("Convert вернула ошибку: %v", err)
	}

	if got := result.Result.FloatString(1); got != "181.0" {
		t.Fatalf("Convert = %s, ожидали 181.0", got)
	}
}

func TestServiceConvertSameCurrency(t *testing.T) {
	service, err := New(currencyRepositoryStub{}, &rateRepositoryStub{})
	if err != nil {
		t.Fatalf("New вернула ошибку: %v", err)
	}

	result, err := service.Convert(context.Background(), "15.25", "rub", "RUB", nil)
	if err != nil {
		t.Fatalf("Convert вернула ошибку: %v", err)
	}

	if got := result.Result.FloatString(2); got != "15.25" {
		t.Fatalf("Convert = %s, ожидали 15.25", got)
	}
}
