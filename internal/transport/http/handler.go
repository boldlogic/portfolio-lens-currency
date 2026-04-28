package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/boldlogic/portfolio-lens-currency/internal/domain/currency"
	"github.com/boldlogic/portfolio-lens-currency/internal/domain/domainerr"
	"github.com/boldlogic/portfolio-lens-currency/internal/domain/rate"
	"github.com/boldlogic/portfolio-lens-currency/internal/domain/source"
	"github.com/boldlogic/portfolio-lens-currency/internal/service"
	"github.com/boldlogic/portfolio-lens-currency/pkg/currencies"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) http.Handler {
	h := &Handler{service: service}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/currencies/", h.getCurrency)
	mux.HandleFunc("GET /api/v1/fx/rates/current", h.getCurrentRate)
	mux.HandleFunc("GET /api/v1/fx/rates", h.getRate)
	mux.HandleFunc("POST /api/v1/fx/rates", h.saveRate)
	mux.HandleFunc("POST /api/v1/fx/convert", h.convert)

	return mux
}

type problemResponse struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

type currencyResponse struct {
	Code        string  `json:"code"`
	NumericCode int16   `json:"numericCode"`
	Name        *string `json:"name,omitempty"`
	LatinName   string  `json:"latinName"`
	MinorUnits  int16   `json:"minorUnits"`
	Active      bool    `json:"active"`
}

type rateResponse struct {
	SourceCode string  `json:"sourceCode"`
	Base       string  `json:"base"`
	Quote      string  `json:"quote"`
	Value      string  `json:"value"`
	ObservedAt string  `json:"observedAt"`
	ValidFrom  string  `json:"validFrom"`
	ValidTo    *string `json:"validTo,omitempty"`
}

type saveRateRequest struct {
	SourceCode string  `json:"sourceCode"`
	Base       string  `json:"base"`
	Quote      string  `json:"quote"`
	Value      string  `json:"value"`
	ObservedAt string  `json:"observedAt"`
	ValidFrom  string  `json:"validFrom"`
	ValidTo    *string `json:"validTo"`
}

type convertRequest struct {
	Amount string  `json:"amount"`
	From   string  `json:"from"`
	To     string  `json:"to"`
	At     *string `json:"at"`
}

type convertResponse struct {
	From   string        `json:"from"`
	To     string        `json:"to"`
	Amount string        `json:"amount"`
	Result string        `json:"result"`
	Rate   *rateResponse `json:"rate,omitempty"`
}

func (h *Handler) getCurrency(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/api/v1/currencies/")
	if code == "" || strings.Contains(code, "/") {
		writeProblem(w, http.StatusNotFound, "not_found", "ресурс не найден")
		return
	}

	currencyModel, err := h.service.GetCurrency(r.Context(), code)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toCurrencyResponse(currencyModel))
}

func (h *Handler) getCurrentRate(w http.ResponseWriter, r *http.Request) {
	fxRate, err := h.service.GetRate(r.Context(), r.URL.Query().Get("base"), r.URL.Query().Get("quote"), nil)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toRateResponse(fxRate))
}

func (h *Handler) getRate(w http.ResponseWriter, r *http.Request) {
	at, err := parseOptionalTime(r.URL.Query().Get("at"))
	if err != nil {
		writeError(w, err)
		return
	}

	fxRate, err := h.service.GetRate(r.Context(), r.URL.Query().Get("base"), r.URL.Query().Get("quote"), at)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toRateResponse(fxRate))
}

func (h *Handler) saveRate(w http.ResponseWriter, r *http.Request) {
	var request saveRateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeProblem(w, http.StatusBadRequest, "bad_request", "некорректный JSON")
		return
	}

	observedAt, err := parseRequiredTime(request.ObservedAt)
	if err != nil {
		writeError(w, err)
		return
	}

	validFrom, err := parseRequiredTime(request.ValidFrom)
	if err != nil {
		writeError(w, err)
		return
	}

	validTo, err := parseOptionalTimePtr(request.ValidTo)
	if err != nil {
		writeError(w, err)
		return
	}

	fxRate, err := h.service.SaveRate(r.Context(), rate.NewRateInput{
		SourceCode: request.SourceCode,
		Base:       request.Base,
		Quote:      request.Quote,
		Value:      request.Value,
		ObservedAt: observedAt,
		ValidFrom:  validFrom,
		ValidTo:    validTo,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toRateResponse(fxRate))
}

func (h *Handler) convert(w http.ResponseWriter, r *http.Request) {
	var request convertRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeProblem(w, http.StatusBadRequest, "bad_request", "некорректный JSON")
		return
	}

	at, err := parseOptionalTimePtr(request.At)
	if err != nil {
		writeError(w, err)
		return
	}

	result, err := h.service.Convert(r.Context(), request.Amount, request.From, request.To, at)
	if err != nil {
		writeError(w, err)
		return
	}

	response := convertResponse{
		From:   result.From.String(),
		To:     result.To.String(),
		Amount: result.Amount.FloatString(18),
		Result: result.Result.FloatString(18),
	}
	if result.Rate.Value != nil {
		rateResponse := toRateResponse(result.Rate)
		response.Rate = &rateResponse
	}

	writeJSON(w, http.StatusOK, response)
}

func toCurrencyResponse(model currency.Currency) currencyResponse {
	return currencyResponse{
		Code:        model.Code.String(),
		NumericCode: model.NumericCode,
		Name:        model.Name,
		LatinName:   model.LatinName,
		MinorUnits:  model.MinorUnits,
		Active:      model.Active,
	}
}

func toRateResponse(model rate.Rate) rateResponse {
	var validTo *string
	if model.ValidTo != nil {
		formatted := model.ValidTo.Format(time.RFC3339)
		validTo = &formatted
	}

	return rateResponse{
		SourceCode: string(model.SourceCode),
		Base:       model.Pair.Base.String(),
		Quote:      model.Pair.Quote.String(),
		Value:      model.Value.FloatString(18),
		ObservedAt: model.ObservedAt.Format(time.RFC3339),
		ValidFrom:  model.ValidFrom.Format(time.RFC3339),
		ValidTo:    validTo,
	}
}

func parseRequiredTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, rate.ErrInvalidRate
	}
	return parsed, nil
}

func parseOptionalTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := parseRequiredTime(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseOptionalTimePtr(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	return parseOptionalTime(*value)
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domainerr.ErrNotFound):
		writeProblem(w, http.StatusNotFound, "not_found", "данные не найдены")
	case errors.Is(err, currencies.ErrWrongISOCharCode),
		errors.Is(err, currencies.ErrNotExistingCurrency),
		errors.Is(err, currency.ErrInvalidCurrency),
		errors.Is(err, rate.ErrInvalidRate),
		errors.Is(err, source.ErrInvalidSource):
		writeProblem(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
	default:
		writeProblem(w, http.StatusInternalServerError, "internal_error", "что-то пошло не так")
	}
}

func writeProblem(w http.ResponseWriter, status int, title, detail string) {
	writeJSON(w, status, problemResponse{
		Title:  title,
		Detail: detail,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
