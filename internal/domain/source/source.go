package source

import (
	"errors"
	"strings"
)

var ErrInvalidSource = errors.New("некорректный источник курса")

type Code string

type Type string

const (
	TypeCBR    Type = "cbr"
	TypeQUIK   Type = "quik"
	TypeMOEX   Type = "moex"
	TypeBroker Type = "broker"
)

type Source struct {
	Code   Code
	Type   Type
	Name   string
	Active bool
}

type PriorityRule struct {
	ID         int64
	BaseCode   string
	QuoteCode  string
	SourceCode Code
	Priority   int16
	Active     bool
}

func NewCode(code string) (Code, error) {
	normalized := strings.ToLower(strings.TrimSpace(code))
	if normalized == "" {
		return "", ErrInvalidSource
	}
	return Code(normalized), nil
}

func NewType(sourceType string) (Type, error) {
	switch Type(strings.ToLower(strings.TrimSpace(sourceType))) {
	case TypeCBR:
		return TypeCBR, nil
	case TypeQUIK:
		return TypeQUIK, nil
	case TypeMOEX:
		return TypeMOEX, nil
	case TypeBroker:
		return TypeBroker, nil
	default:
		return "", ErrInvalidSource
	}
}

func NewSource(code, sourceType, name string, active bool) (Source, error) {
	normalizedCode, err := NewCode(code)
	if err != nil {
		return Source{}, err
	}

	normalizedType, err := NewType(sourceType)
	if err != nil {
		return Source{}, err
	}

	if strings.TrimSpace(name) == "" {
		return Source{}, ErrInvalidSource
	}

	return Source{
		Code:   normalizedCode,
		Type:   normalizedType,
		Name:   strings.TrimSpace(name),
		Active: active,
	}, nil
}
