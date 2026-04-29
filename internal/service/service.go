package service

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/boldlogic/portfolio-lens-currency/internal/domain/currency"
	"github.com/boldlogic/portfolio-lens-currency/internal/domain/domainerr"
	"github.com/boldlogic/portfolio-lens-currency/internal/domain/rate"
	"github.com/boldlogic/portfolio-lens-currency/pkg/currencies"
)

var (
	ErrNotFound       = domainerr.ErrNotFound
	ErrInvalidService = errors.New("некорректная конфигурация сервиса")
)

type CurrencyRepository interface {
	GetCurrency(ctx context.Context, code currencies.CurrencyCode) (currency.Currency, error)
}

type RateRepository interface {
	GetRate(ctx context.Context, query rate.Query) (rate.Rate, error)
	SaveRate(ctx context.Context, fxRate rate.Rate) error
}

type Service struct {
	currencies CurrencyRepository
	rates      RateRepository
}

type ConversionResult struct {
	From   currencies.CurrencyCode
	To     currencies.CurrencyCode
	Amount *big.Rat
	Result *big.Rat
	Rate   rate.Rate
}

func New(currencies CurrencyRepository, rates RateRepository) (*Service, error) {
	if currencies == nil || rates == nil {
		return nil, ErrInvalidService
	}
	return &Service{
		currencies: currencies,
		rates:      rates,
	}, nil
}

func (s *Service) GetCurrency(ctx context.Context, code string) (currency.Currency, error) {
	currencyCode, err := currencies.ParseCurrencyCode(code)
	if err != nil {
		return currency.Currency{}, err
	}
	return s.currencies.GetCurrency(ctx, currencyCode)
}

func (s *Service) GetRate(ctx context.Context, base, quote string, at *time.Time) (rate.Rate, error) {
	query, err := rate.NewQuery(base, quote, at)
	if err != nil {
		return rate.Rate{}, err
	}
	return s.rates.GetRate(ctx, query)
}

func (s *Service) SaveRate(ctx context.Context, input rate.NewRateInput) (rate.Rate, error) {
	fxRate, err := rate.NewRate(input)
	if err != nil {
		return rate.Rate{}, err
	}

	if err := s.rates.SaveRate(ctx, fxRate); err != nil {
		return rate.Rate{}, err
	}

	return fxRate, nil
}

func (s *Service) Convert(ctx context.Context, amount, from, to string, at *time.Time) (ConversionResult, error) {
	value, ok := new(big.Rat).SetString(amount)
	if !ok || value.Sign() < 0 {
		return ConversionResult{}, fmt.Errorf("проверить сумму: %w", rate.ErrInvalidRate)
	}

	fromCode, err := currencies.ParseCurrencyCode(from)
	if err != nil {
		return ConversionResult{}, err
	}

	toCode, err := currencies.ParseCurrencyCode(to)
	if err != nil {
		return ConversionResult{}, err
	}

	if fromCode == toCode {
		return ConversionResult{
			From:   fromCode,
			To:     toCode,
			Amount: value,
			Result: new(big.Rat).Set(value),
		}, nil
	}

	fxRate, err := s.GetRate(ctx, fromCode.String(), toCode.String(), at)
	if err != nil {
		return ConversionResult{}, err
	}

	return ConversionResult{
		From:   fromCode,
		To:     toCode,
		Amount: value,
		Result: new(big.Rat).Mul(value, fxRate.Value),
		Rate:   fxRate,
	}, nil
}
