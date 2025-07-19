package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/storage/inmem"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

//go:generate go tool mockgen -package=mocks -destination=../../mocks/service/authz.go -source=authz.go AuthzServiceCache

// AuthzServiceCache is the interface for the authz service cache methods.
// This must be used in other services that need to access the authz cache key and invalidate it.
// implement as dependency injection in the AuthzServiceCache interface.
// NOTE: This is defined here to allows other services to use the AuthzServiceCache interface
// without importing the entire authz service package.
type AuthzServiceCache interface {
	GetUserAuthzCacheKey(userID uuid.UUID) string
	InvalidateUserAuthzCache(userID uuid.UUID)
}

type AuthzServiceConf struct {
	Repository    UsersRepository
	CacheService  *CacheService
	RegoQuery     string
	RegoPolicy    string
	OT            *o11y.OpenTelemetry
	MetricsPrefix string
}

type authzServiceMetrics struct {
	serviceCalls metric.Int64Counter
}

type AuthzService struct {
	repository    UsersRepository
	cacheService  *CacheService
	regoQuery     string
	regoPolicy    string
	ot            *o11y.OpenTelemetry
	metricsPrefix string
	metrics       authzServiceMetrics
}

func NewAuthzService(conf AuthzServiceConf) (*AuthzService, error) {
	if conf.Repository == nil {
		return nil, &model.InvalidRepositoryError{Message: "Repository is nil, but it is required for AuthzService"}
	}

	if conf.OT == nil {
		return nil, &model.InvalidOTConfigurationError{Message: "OpenTelemetry is nil, but it is required for AuthzService"}
	}

	if conf.RegoQuery == "" {
		return nil, &model.InvalidRegoQueryError{Message: "rego query cannot be empty"}
	}

	if conf.RegoPolicy == "" {
		return nil, &model.InvalidRegoPolicyError{Message: "rego policy cannot be empty"}
	}

	ref := &AuthzService{
		repository:   conf.Repository,
		cacheService: conf.CacheService,
		regoQuery:    conf.RegoQuery,
		regoPolicy:   conf.RegoPolicy,
		ot:           conf.OT,
	}
	if conf.MetricsPrefix != "" {
		ref.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		ref.metricsPrefix += "_"
	}

	serviceCalls, err := ref.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", ref.metricsPrefix, "services_calls_total"),
		metric.WithDescription("The number of calls to the authz service"),
	)
	if err != nil {
		return nil, err
	}

	ref.metrics.serviceCalls = serviceCalls

	return ref, nil
}

func (ref *AuthzService) IsAuthorized(ctx context.Context, userID uuid.UUID, requestAction, requestResource string) (bool, error) {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Authz.IsAuthorized")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("action", requestAction),
		attribute.String("resource", requestResource),
	)

	if userID == uuid.Nil {
		errorType := &model.InvalidUserIDError{Message: "user ID cannot be empty"}
		return false, o11y.RecordError(ctx, span, errorType, ref.metrics.serviceCalls, metricCommonAttributes, "service.Authz.IsAuthorized")
	}

	var userAuth map[string]any
	var err error

	// Get the user Authz from the database, cache is disabled
	if ref.cacheService.cache == nil {
		slog.Debug("service.Authz.IsAuthorized", "cache", "disabled")
		// Get the user Authz from the repository
		userAuth, err = ref.repository.SelectAuthz(ctx, userID)
		if err != nil {
			return false, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Authz.IsAuthorized")
		}
	} else {
		cacheKey := fmt.Sprintf("authz:%s", userID.String())
		slog.Debug("service.Authz.IsAuthorized", "cache", "enabled", "key", cacheKey)
		ctxCache, cancel := context.WithTimeout(ctx, ref.cacheService.queryTimeout)
		defer cancel()

		userAuth, err = FromCacheOrDB(
			ctxCache,
			ref.cacheService.cache,
			cacheKey,
			CacheEncoderTypeGob,
			func() (map[string]any, error) {
				return ref.repository.SelectAuthz(ctx, userID)
			},
			ref.cacheService.queryTimeout,
			ref.cacheService.entitiesTTL,
		)
		if err != nil {
			return false, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Authz.IsAuthorized")
		}
	}

	slog.Debug("service.Authz.IsAuthorized", "permissions", userAuth, "userID", userID.String())

	opaInput := map[string]any{
		"user_id":  userID.String(),
		"action":   requestAction,
		"resource": requestResource,
	}

	// Manually create the storage layer. inmem.NewFromObject returns an
	// in-memory store containing the supplied data.
	store := inmem.NewFromObject(userAuth)

	query, err := rego.New(
		rego.Query(ref.regoQuery),
		rego.Module("policy.rego", ref.regoPolicy),
		rego.Input(opaInput),
		rego.Store(store),
		rego.EnablePrintStatements(true),
	).PrepareForEval(ctx)
	if err != nil {
		return false, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Authz.IsAuthorized")
	}

	results, err := query.Eval(ctx)
	slog.Debug("service.Authz.IsAuthorized", "results", results)
	if err != nil {
		errorVar := &model.UnauthorizedError{Message: err.Error()}
		return false, o11y.RecordError(ctx, span, errorVar, ref.metrics.serviceCalls, metricCommonAttributes, "service.Authz.IsAuthorized")
	} else if len(results) == 0 {
		errorVar := &model.UnauthorizedError{Message: "unauthorized: no results found"}
		return false, o11y.RecordError(ctx, span, errorVar, ref.metrics.serviceCalls, metricCommonAttributes, "service.Authz.IsAuthorized")
	} else if len(results[0].Expressions) == 0 {
		errorVar := &model.UnauthorizedError{Message: "unauthorized: no expressions found"}
		return false, o11y.RecordError(ctx, span, errorVar, ref.metrics.serviceCalls, metricCommonAttributes, "service.Authz.IsAuthorized")
	}

	isAuthorized := results[0].Expressions[0].Value.(bool)
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "Authorization check completed", attribute.Bool("authorized", isAuthorized))
	return isAuthorized, nil
}

// Helper functions for common patterns

// setupContext creates a context with a span and common attributes for metrics.
// Returns the new context, span, and common metric attributes.
func (ref *AuthzService) setupContext(ctx context.Context, operation string) (context.Context, trace.Span, []attribute.KeyValue) {
	ctx, span := ref.ot.Traces.Tracer.Start(ctx, operation)

	span.SetAttributes(
		attribute.String("component", operation),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", operation),
	}

	return ctx, span, metricCommonAttributes
}

// GetUserAuthzCacheKey generates a cache key for user authorization based on userID.
func (ref *AuthzService) GetUserAuthzCacheKey(userID uuid.UUID) string {
	return fmt.Sprintf("authz:%s", userID)
}

// InvalidateUserAuthzCache removes the user authorization cache entry for the given userID.
func (ref *AuthzService) InvalidateUserAuthzCache(userID uuid.UUID) {
	if userID == uuid.Nil {
		return
	}

	cacheKey := ref.GetUserAuthzCacheKey(userID)
	if ref.cacheService != nil && ref.cacheService.cache != nil {
		slog.Debug("service.Authz.InvalidateUserAuthzCache", "cache", "removing cache", "key", cacheKey)
		ref.cacheService.Remove(context.Background(), cacheKey)
	}
}

