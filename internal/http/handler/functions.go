package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"github.com/p2p-b2b/qfv"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// setupContext creates a context with common attributes for tracing and metrics.
// Returns the context, span, and common metric attributes for consistent tracking across handlers.
func setupContext(r *http.Request, tracer trace.Tracer, component string) (context.Context, trace.Span, []attribute.KeyValue) {
	ctx, span := tracer.Start(r.Context(), component)

	// Get the base path by removing the last part after the last slash
	path := r.URL.Path
	lastSlashIndex := strings.LastIndex(path, "/")
	if lastSlashIndex != -1 {
		path = path[:lastSlashIndex]
	}

	span.SetAttributes(
		attribute.String("component", component),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", path),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", component),
		attribute.String("http.method", r.Method),
		attribute.String("http.path", path),
	}

	return ctx, span, metricCommonAttributes
}

// recordError records an error in the span, logs it, and updates metrics with failure status.
// It is a wrapper around o11y.RecordError with HTTP specific context
// Returns the error for chainable error handling.
func recordError(
	ctx context.Context,
	span trace.Span,
	err error,
	metricsCounter metric.Int64Counter,
	metricAttrs []attribute.KeyValue,
	statusCode int,
	component string,
	details ...any,
) error {
	// Add HTTP status code to metric attributes
	metricAttrs = append(metricAttrs, attribute.String("code", fmt.Sprintf("%d", statusCode)))

	return o11y.RecordError(
		ctx,
		span,
		err,
		metricsCounter,
		metricAttrs,
		component,
		details...,
	)
}

// recordSuccess records a successful operation in the span and updates metrics with success status.
// It is a wrapper around o11y.RecordSuccess with HTTP specific context.
func recordSuccess(
	ctx context.Context,
	span trace.Span,
	metricsCounter metric.Int64Counter,
	metricAttrs []attribute.KeyValue,
	statusCode int,
	message string,
	attrs ...attribute.KeyValue,
) {
	// Add HTTP status code to metric attributes
	metricAttrs = append(metricAttrs, attribute.String("code", fmt.Sprintf("%d", statusCode)))

	o11y.RecordSuccess(
		ctx,
		span,
		metricsCounter,
		metricAttrs,
		message,
		attrs...,
	)
}

// parseUUIDQueryParams parses a string into a UUID.
// If the input is empty, it returns an error.
// If the input is not a valid UUID, it returns an error.
// If the input is a nil UUID, it returns an error.
func parseUUIDQueryParams(input string) (uuid.UUID, error) {
	if input == "" {
		return uuid.Nil, &InvalidUUIDError{Message: "input is empty"}
	}

	id, err := uuid.Parse(input)
	if err != nil {
		return uuid.Nil, &InvalidUUIDError{UUID: input, Message: err.Error()}
	}

	if id == uuid.Nil {
		return uuid.Nil, &InvalidUUIDError{UUID: input, Message: "input is nil"}
	}

	return id, nil
}

// parseSortQueryParams parses a string into a sort field.
func parseSortQueryParams(sort string, allowedFields []string) (string, error) {
	if sort == "" {
		return "", nil
	}

	_, err := qfv.NewSortParser(allowedFields).Parse(sort)
	if err != nil {
		return "", err
	}

	return sort, nil
}

// parseFilterQueryParams parses a string into a filter field.
func parseFilterQueryParams(filter string, allowedFields []string) (string, error) {
	if filter == "" {
		return "", nil
	}

	_, err := qfv.NewFilterParser(allowedFields).Parse(filter)
	if err != nil {
		return "", err
	}

	return filter, nil
}

// parseFieldsQueryParams parses a string into a list of fields.
func parseFieldsQueryParams(fields string, allowedFields []string) (string, error) {
	if fields == "" {
		return "", nil
	}

	_, err := qfv.NewFieldsParser(allowedFields).Parse(fields)
	if err != nil {
		return "", err
	}

	ret := strings.ReplaceAll(fields, " ", "")

	return ret, nil
}

// parseNextTokenQueryParams parses a string into a nextToken field.
func parseNextTokenQueryParams(nextToken string) (string, error) {
	if nextToken != "" {
		_, _, _, err := model.DecodeToken(nextToken, model.TokenDirectionNext)
		if err != nil {
			return "", &model.InvalidPaginatorTokenError{Message: err.Error()}
		}
	}

	return nextToken, nil
}

// parsePrevTokenQueryParams parses a string into a prevToken field.
func parsePrevTokenQueryParams(prevToken string) (string, error) {
	if prevToken != "" {
		_, _, _, err := model.DecodeToken(prevToken, model.TokenDirectionPrev)
		if err != nil {
			return "", &model.InvalidPaginatorTokenError{Message: err.Error()}
		}
	}

	return prevToken, nil
}

// parseLimitQueryParams parses a string into a limit field.
func parseLimitQueryParams(limit string) (int, error) {
	var limitInt int
	var err error

	if limit == "" {
		return model.PaginatorDefaultLimit, nil
	}

	// check if this is a valid integer
	if limitInt, err = strconv.Atoi(limit); err != nil {
		return 0, &model.InvalidPaginatorLimitError{MinLimit: model.PaginatorMinLimit, MaxLimit: model.PaginatorMaxLimit}
	}

	if limitInt == 0 {
		limitInt = model.PaginatorDefaultLimit
	}

	if limitInt < model.PaginatorMinLimit {
		return 0, &model.InvalidPaginatorLimitError{MinLimit: model.PaginatorMinLimit, MaxLimit: model.PaginatorMaxLimit}
	} else if limitInt > model.PaginatorMaxLimit {
		limitInt = model.PaginatorMaxLimit
	}

	return limitInt, nil
}

// parseListQueryParams parses a list of strings into a list of UUIDs.
func parseListQueryParams(params map[string]any, fieldsFields, filterFields, sortFields []string) (
	sort string,
	filter string,
	fields string,
	nextToken string,
	prevToken string,
	limit int,
	err error,
) {
	sort, err = parseSortQueryParams(params["sort"].(string), sortFields)
	if err != nil {
		return "", "", "", "", "", 0, err
	}

	filter, err = parseFilterQueryParams(params["filter"].(string), filterFields)
	if err != nil {
		return "", "", "", "", "", 0, err
	}

	fields, err = parseFieldsQueryParams(params["fields"].(string), fieldsFields)
	if err != nil {
		return "", "", "", "", "", 0, err
	}

	nextToken, err = parseNextTokenQueryParams(params["nextToken"].(string))
	if err != nil {
		return "", "", "", "", "", 0, err
	}

	prevToken, err = parsePrevTokenQueryParams(params["prevToken"].(string))
	if err != nil {
		return "", "", "", "", "", 0, err
	}

	limit, err = parseLimitQueryParams(params["limit"].(string))
	if err != nil {
		return "", "", "", "", "", 0, err
	}

	return sort, filter, fields, nextToken, prevToken, limit, nil
}
