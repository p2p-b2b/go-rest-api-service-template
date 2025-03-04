package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"go.opentelemetry.io/otel/metric"
)

type HealthRepositoryConfig struct {
	DB             *pgxpool.Pool
	MaxPingTimeout time.Duration
	OT             *o11y.OpenTelemetry
	MetricsPrefix  string
}

type healthRepositoryMetrics struct {
	repositoryCalls metric.Int64Counter
}

// this implement repository.HealthRepository
// HealthRepository is a PostgreSQL store.
type HealthRepository struct {
	db             *pgxpool.Pool
	maxPingTimeout time.Duration
	ot             *o11y.OpenTelemetry
	metricsPrefix  string
	metrics        healthRepositoryMetrics
}

// NewHealthRepository creates a new HealthRepository.
func NewHealthRepository(conf HealthRepositoryConfig) (*HealthRepository, error) {
	if conf.DB == nil {
		return nil, ErrDBInvalidConfiguration
	}

	if conf.MaxPingTimeout < 10*time.Millisecond {
		return nil, ErrDBInvalidMaxPingTimeout
	}

	if conf.OT == nil {
		return nil, ErrOTInvalidConfiguration
	}

	repo := &HealthRepository{
		db:             conf.DB,
		maxPingTimeout: conf.MaxPingTimeout,
		ot:             conf.OT,
	}

	if conf.MetricsPrefix != "" {
		repo.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		repo.metricsPrefix += "_"
	}

	repositoryCalls, err := repo.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", repo.metricsPrefix, "repository_calls_total"),
		metric.WithDescription("The number of calls to the health repository"),
	)
	if err != nil {
		slog.Error("repository.Health.NewHealthRepository", "error", err)
		return nil, err
	}

	repo.metrics.repositoryCalls = repositoryCalls

	return repo, nil
}

// DriverName returns the name of the driver.
func (ref *HealthRepository) DriverName() string {
	return sql.Drivers()[0]
}

// PingContext verifies a connection to the repository is still alive, establishing a connection if necessary.
func (ref *HealthRepository) PingContext(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, ref.maxPingTimeout)
	defer cancel()

	ctx, span := ref.ot.Traces.Tracer.Start(ctx, "repository.Health.PingContext")
	defer span.End()

	return ref.db.Ping(ctx)
}
