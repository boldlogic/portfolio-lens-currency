package rate

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/boldlogic/portfolio-lens-currency/internal/domain/source"
	"github.com/boldlogic/portfolio-lens-currency/pkg/currencies"
)

var ErrInvalidRate = errors.New("некорректный курс валют")

type Rate struct {
	SourceCode source.Code
	Pair       currencies.CurrencyPair
	Value      *big.Rat
	ObservedAt time.Time
	ValidFrom  time.Time
	ValidTo    *time.Time
}

type Query struct {
	Pair currencies.CurrencyPair
	At   *time.Time
}

type NewRateInput struct {
	SourceCode string
	Base       string
	Quote      string
	Value      string
	ObservedAt time.Time
	ValidFrom  time.Time
	ValidTo    *time.Time
}

func NewRate(input NewRateInput) (Rate, error) {
	sourceCode, err := source.NewCode(input.SourceCode)
	if err != nil {
		return Rate{}, fmt.Errorf("проверить источник курса: %w", err)
	}

	pair, err := currencies.NewCurrencyPair(input.Base, input.Quote)
	if err != nil {
		return Rate{}, err
	}

	value, err := NewValue(input.Value)
	if err != nil {
		return Rate{}, err
	}

	if input.ObservedAt.IsZero() || input.ValidFrom.IsZero() {
		return Rate{}, ErrInvalidRate
	}
	if input.ValidTo != nil && input.ValidTo.Before(input.ValidFrom) {
		return Rate{}, ErrInvalidRate
	}

	return Rate{
		SourceCode: sourceCode,
		Pair:       pair,
		Value:      value,
		ObservedAt: input.ObservedAt,
		ValidFrom:  input.ValidFrom,
		ValidTo:    input.ValidTo,
	}, nil
}

func NewValue(value string) (*big.Rat, error) {
	rat, ok := new(big.Rat).SetString(value)
	if !ok || rat.Sign() <= 0 {
		return nil, ErrInvalidRate
	}
	return rat, nil
}

func NewQuery(base, quote string, at *time.Time) (Query, error) {
	pair, err := currencies.NewCurrencyPair(base, quote)
	if err != nil {
		return Query{}, err
	}
	return Query{
		Pair: pair,
		At:   at,
	}, nil
}
