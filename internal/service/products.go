package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

//go:generate go tool mockgen -package=mocks -destination=../../mocks/service/products.go -source=products.go ProductsRepository

// ProductsRepository is the interface for the products repository methods.
type ProductsRepository interface {
	SelectByIDByProjectID(ctx context.Context, id uuid.UUID, projectID uuid.UUID) (*model.Product, error)
	SelectByProjectID(ctx context.Context, projectID uuid.UUID, input *model.SelectProductsInput) (*model.SelectProductsOutput, error)
	Select(ctx context.Context, input *model.SelectProductsInput) (*model.SelectProductsOutput, error)

	Insert(ctx context.Context, input *model.InsertProductInput) error
	Update(ctx context.Context, input *model.UpdateProductInput) error
	Delete(ctx context.Context, input *model.DeleteProductInput) error

	LinkToPaymentProcessor(ctx context.Context, input *model.LinkProductToPaymentProcessorInput) error
	UnlinkFromPaymentProcessor(ctx context.Context, input *model.UnlinkProductFromPaymentProcessorInput) error
}

type ProductsServiceConf struct {
	Repository    ProductsRepository
	OT            *o11y.OpenTelemetry
	MetricsPrefix string
}

type productsServiceMetrics struct {
	serviceCalls metric.Int64Counter
}

type ProductsService struct {
	repository    ProductsRepository
	ot            *o11y.OpenTelemetry
	metricsPrefix string
	metrics       productsServiceMetrics
}

// NewProductsService creates a new ProductsService.
func NewProductsService(conf ProductsServiceConf) (*ProductsService, error) {
	if conf.Repository == nil {
		return nil, &model.InvalidRepositoryError{Message: "Repository is nil, but it is required for ProductsService"}
	}

	if conf.OT == nil {
		return nil, &model.InvalidOTConfigurationError{Message: "OpenTelemetry is nil, but it is required for ProductsService"}
	}

	service := &ProductsService{
		repository: conf.Repository,
		ot:         conf.OT,
	}

	if conf.MetricsPrefix != "" {
		service.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		service.metricsPrefix += "_"
	}

	serviceCalls, err := service.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", service.metricsPrefix, "services_calls_total"),
		metric.WithDescription("The number of calls to the products service"),
	)
	if err != nil {
		return nil, err
	}

	service.metrics.serviceCalls = serviceCalls

	return service, nil
}

func (ref *ProductsService) CreateByProjectID(ctx context.Context, input *model.CreateProductInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Products.CreateByProjectID")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.CreateByProjectID")
	}

	if input.ID == uuid.Nil {
		var err error
		input.ID, err = uuid.NewV7()
		if err != nil {
			return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.CreateByProjectID")
		}
	}

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.CreateByProjectID")
	}

	if err := ref.repository.Insert(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.CreateByProjectID")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "product created successfully",
		attribute.String("product.id", input.ID.String()))

	return nil
}

func (ref *ProductsService) UpdateByIDByProjectID(ctx context.Context, input *model.UpdateProductInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Products.UpdateByIDByProjectID")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.UpdateByIDByProjectID")
	}

	span.SetAttributes(
		attribute.String("product.id", input.ID.String()),
	)

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.UpdateByIDByProjectID")
	}

	if err := ref.repository.Update(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.UpdateByIDByProjectID")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "product updated successfully",
		attribute.String("product.id", input.ID.String()))

	return nil
}

func (ref *ProductsService) DeleteByIDByProjectID(ctx context.Context, input *model.DeleteProductInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Products.DeleteByIDByProjectID")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.DeleteByIDByProjectID")
	}

	span.SetAttributes(
		attribute.String("product.id", input.ID.String()),
	)

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.DeleteByIDByProjectID")
	}

	if err := ref.repository.Delete(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.DeleteByIDByProjectID")
	}

	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "product deleted successfully",
		attribute.String("product.id", input.ID.String()))

	return nil
}

func (ref *ProductsService) GetByIDByProjectID(ctx context.Context, projectID uuid.UUID, id uuid.UUID) (*model.Product, error) {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Products.GetByIDByProjectID")
	defer span.End()

	if id == uuid.Nil {
		invalidErr := &model.InvalidProductIDError{Message: "invalid product ID. It is nil"}
		return nil, o11y.RecordError(ctx, span, invalidErr, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.GetByIDByProjectID")
	}

	if projectID == uuid.Nil {
		invalidErr := &model.InvalidProjectIDError{Message: "invalid project ID. It is nil"}
		return nil, o11y.RecordError(ctx, span, invalidErr, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.GetByIDByProjectID")
	}

	span.SetAttributes(attribute.String("products.id", id.String()))

	out, err := ref.repository.SelectByIDByProjectID(ctx, id, projectID)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.GetByIDByProjectID")
	}

	slog.Debug("service.Products.GetByIDByProjectID", "product.id", out.ID)
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "product found successfully", attribute.String("product.id", out.ID.String()))

	return out, nil
}

func (ref *ProductsService) ListByProjectID(ctx context.Context, projectID uuid.UUID, input *model.ListProductsInput) (*model.ListProductsOutput, error) {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Products.ListByProjectID")
	defer span.End()

	if projectID == uuid.Nil {
		invalidErr := &model.InvalidProjectIDError{Message: "invalid project ID. It is nil"}
		return nil, o11y.RecordError(ctx, span, invalidErr, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.ListByProjectID")
	}

	if input == nil {
		input = &model.ListProductsInput{}
	}

	span.SetAttributes(
		attribute.String("project.id", projectID.String()),
		attribute.String("sort", input.Sort),
		attribute.String("fields", input.Fields),
		attribute.String("filter", input.Filter),
		attribute.Int("limit", input.Paginator.Limit),
	)

	if err := input.Validate(); err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.ListByProjectID")
	}

	out, err := ref.repository.SelectByProjectID(ctx, projectID, input)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.ListByProjectID")
	}

	slog.Debug("service.Products.ListByProjectID", "models", len(out.Items))
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "products listed successfully",
		attribute.Int("count", len(out.Items)))

	return out, nil
}

func (ref *ProductsService) List(ctx context.Context, input *model.ListProductsInput) (*model.ListProductsOutput, error) {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Products.List")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return nil, o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.List")
	}

	span.SetAttributes(
		attribute.String("sort", input.Sort),
		attribute.String("fields", input.Fields),
		attribute.String("filter", input.Filter),
		attribute.Int("limit", input.Paginator.Limit),
	)

	if err := input.Validate(); err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.List")
	}

	out, err := ref.repository.Select(ctx, input)
	if err != nil {
		return nil, o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.List")
	}

	slog.Debug("service.Products.List", "models", len(out.Items))
	o11y.RecordSuccess(ctx, span, ref.metrics.serviceCalls, metricCommonAttributes, "products listed successfully",
		attribute.Int("count", len(out.Items)))

	return out, nil
}

// Helper functions for common patterns

// setupContext creates a context with a span and common attributes for metrics.
// Returns the new context, span, and common metric attributes.
func (ref *ProductsService) setupContext(ctx context.Context, operation string) (context.Context, trace.Span, []attribute.KeyValue) {
	ctx, span := ref.ot.Traces.Tracer.Start(ctx, operation)

	span.SetAttributes(
		attribute.String("component", operation),
	)

	metricCommonAttributes := []attribute.KeyValue{
		attribute.String("component", operation),
	}

	return ctx, span, metricCommonAttributes
}

func (ref *ProductsService) LinkToPaymentProcessor(ctx context.Context, input *model.LinkProductToPaymentProcessorInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Products.LinkToPaymentProcessor")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.LinkToPaymentProcessor")
	}

	span.SetAttributes(
		attribute.String("product.id", input.ProductID.String()),
	)

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.LinkToPaymentProcessor")
	}

	if err := ref.repository.LinkToPaymentProcessor(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.LinkToPaymentProcessor")
	}

	return nil
}

func (ref *ProductsService) UnlinkFromPaymentProcessor(ctx context.Context, input *model.UnlinkProductFromPaymentProcessorInput) error {
	ctx, span, metricCommonAttributes := ref.setupContext(ctx, "service.Products.UnlinkFromPaymentProcessor")
	defer span.End()

	if input == nil {
		errorValue := &model.InvalidInputError{Message: "input is nil"}
		return o11y.RecordError(ctx, span, errorValue, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.UnlinkFromPaymentProcessor")
	}

	span.SetAttributes(
		attribute.String("product.id", input.ProductID.String()),
	)

	if err := input.Validate(); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.UnlinkFromPaymentProcessor")
	}

	if err := ref.repository.UnlinkFromPaymentProcessor(ctx, input); err != nil {
		return o11y.RecordError(ctx, span, err, ref.metrics.serviceCalls, metricCommonAttributes, "service.Products.UnlinkFromPaymentProcessor")
	}

	return nil
}
