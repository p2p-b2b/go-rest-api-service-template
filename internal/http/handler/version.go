package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/middleware"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/respond"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/version"
	"go.opentelemetry.io/otel/metric"
)

// VersionHandlerConf represents the configuration for the version handler.
type VersionHandlerConf struct {
	OT            *o11y.OpenTelemetry
	MetricsPrefix string
}

type versionHandlerMetrics struct {
	handlerCalls metric.Int64Counter
}

// VersionHandler represents the handler for the version of the service.
type VersionHandler struct {
	ot            *o11y.OpenTelemetry
	metricsPrefix string
	metrics       versionHandlerMetrics
}

// NewVersionHandler returns a new instance of VersionHandler.
func NewVersionHandler(conf VersionHandlerConf) (*VersionHandler, error) {
	if conf.OT == nil {
		return nil, &model.InvalidOTConfigurationError{Message: "OpenTelemetry is not configured"}
	}

	handler := &VersionHandler{
		ot: conf.OT,
	}

	if conf.MetricsPrefix != "" {
		handler.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		handler.metricsPrefix += "_"
	}

	handlerCalls, err := handler.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", handler.metricsPrefix, "handlers_calls_total"),
		metric.WithDescription("The number of calls to the version handler"),
	)
	if err != nil {
		return nil, err
	}

	handler.metrics.handlerCalls = handlerCalls

	return handler, nil
}

// RegisterRoutes registers the routes for the version of the service.
func (ref *VersionHandler) RegisterRoutes(mux *http.ServeMux, middlewares ...middleware.Middleware) {
	mdw := middleware.Chain(middlewares...)

	mux.Handle("GET /version", mdw.ThenFunc(ref.get))
}

// get returns the version of the service
//
//	@ID				019791cc-06c7-7eff-a1df-fbc2ad0b27c9
//	@Summary		Get version
//	@Description	Retrieve the current version and build information of the service
//	@Tags			Version
//	@Produce		json
//	@Success		200	{object}	model.Version
//	@Failure		500	{object}	model.HTTPMessage
//	@Router			/version [get]
func (ref *VersionHandler) get(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Version.get")
	defer span.End()

	w.Header().Set("Content-Type", "application/json")

	out := model.Version{
		Version:       version.Version,
		BuildDate:     version.BuildDate,
		GitCommit:     version.GitCommit,
		GitBranch:     version.GitBranch,
		GoVersion:     version.GoVersion,
		GoVersionArch: version.GoVersionArch,
		GoVersionOS:   version.GoVersionOS,
	}

	if err := respond.WriteJSONData(w, http.StatusOK, out); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Version.get")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}
}
