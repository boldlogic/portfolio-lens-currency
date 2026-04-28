package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/boldlogic/portfolio-lens-currency/internal/domain/currency"
	"github.com/boldlogic/portfolio-lens-currency/internal/domain/domainerr"
	"github.com/boldlogic/portfolio-lens-currency/internal/domain/rate"
	"github.com/boldlogic/portfolio-lens-currency/pkg/currencies"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) (*Repository, error) {
	if db == nil {
		return nil, errors.New("подключение к PostgreSQL не задано")
	}
	return &Repository{db: db}, nil
}

func (r *Repository) GetCurrency(ctx context.Context, code currencies.CurrencyCode) (currency.Currency, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT code, numeric_code, name, latin_name, minor_units, active
		FROM currencies
		WHERE code = $1
	`, code.String())

	var (
		rawCode    string
		numeric    int16
		name       sql.NullString
		latinName  string
		minorUnits int16
		active     bool
	)

	if err := row.Scan(&rawCode, &numeric, &name, &latinName, &minorUnits, &active); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return currency.Currency{}, domainerr.ErrNotFound
		}
		return currency.Currency{}, fmt.Errorf("получить валюту: %w", err)
	}

	var namePtr *string
	if name.Valid {
		namePtr = &name.String
	}

	return currency.NewCurrency(currency.NewCurrencyInput{
		Code:        rawCode,
		NumericCode: numeric,
		Name:        namePtr,
		LatinName:   latinName,
		MinorUnits:  minorUnits,
		Active:      active,
	})
}

func (r *Repository) GetRate(ctx context.Context, query rate.Query) (rate.Rate, error) {
	at := any(nil)
	if query.At != nil {
		at = *query.At
	}

	row := r.db.QueryRowContext(ctx, `
		SELECT r.source_code, r.base_code, r.quote_code, r.value::text, r.observed_at, r.valid_from, r.valid_to
		FROM fx_rates r
		JOIN fx_sources s ON s.code = r.source_code
		LEFT JOIN source_priority_rules pr ON pr.active = true
			AND pr.source_code = r.source_code
			AND (pr.base_code IS NULL OR pr.base_code = r.base_code)
			AND (pr.quote_code IS NULL OR pr.quote_code = r.quote_code)
		WHERE r.base_code = $1
			AND r.quote_code = $2
			AND s.active = true
			AND (
				$3::timestamptz IS NULL
				OR (r.valid_from <= $3 AND (r.valid_to IS NULL OR r.valid_to > $3))
			)
			AND ($3::timestamptz IS NOT NULL OR r.valid_to IS NULL)
		ORDER BY COALESCE(pr.priority, 32767), r.observed_at DESC
		LIMIT 1
	`, query.Pair.Base.String(), query.Pair.Quote.String(), at)

	return scanRate(row)
}

func (r *Repository) SaveRate(ctx context.Context, fxRate rate.Rate) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO fx_rates (
			source_code,
			base_code,
			quote_code,
			value,
			observed_at,
			valid_from,
			valid_to
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (source_code, base_code, quote_code, observed_at)
		DO UPDATE SET
			value = EXCLUDED.value,
			valid_from = EXCLUDED.valid_from,
			valid_to = EXCLUDED.valid_to
	`, fxRate.SourceCode, fxRate.Pair.Base.String(), fxRate.Pair.Quote.String(), fxRate.Value.FloatString(18), fxRate.ObservedAt, fxRate.ValidFrom, fxRate.ValidTo)
	if err != nil {
		return fmt.Errorf("сохранить курс валют: %w", err)
	}
	return nil
}

type rateScanner interface {
	Scan(dest ...any) error
}

func scanRate(row rateScanner) (rate.Rate, error) {
	var (
		sourceCode string
		base       string
		quote      string
		value      string
		observedAt time.Time
		validFrom  time.Time
		validTo    sql.NullTime
	)

	if err := row.Scan(&sourceCode, &base, &quote, &value, &observedAt, &validFrom, &validTo); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return rate.Rate{}, domainerr.ErrNotFound
		}
		return rate.Rate{}, fmt.Errorf("получить курс валют: %w", err)
	}

	var validToPtr *time.Time
	if validTo.Valid {
		validToPtr = &validTo.Time
	}

	return rate.NewRate(rate.NewRateInput{
		SourceCode: sourceCode,
		Base:       base,
		Quote:      quote,
		Value:      value,
		ObservedAt: observedAt,
		ValidFrom:  validFrom,
		ValidTo:    validToPtr,
	})
}
