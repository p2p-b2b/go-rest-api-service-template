package config

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// Default Database Configuration
	DefaultMetricsPath = "metrics"
	DefaultNamespace   = "namespace"
)

type MetricsConfig struct {
	MetricsPath Field[string]
	Namespace   Field[string]
	AppName     string
}

type Metrics struct {

	//Handler
	handler http.Handler

	//namespace
	namespace string

	// Global
	Up prometheus.Gauge

	// Calls
	Http_calls prometheus.CounterVec
}

func NewMetricsConfig(appName string) *MetricsConfig {
	return &MetricsConfig{
		MetricsPath: NewField("metrics.path", "METRICS_PATH", "path to scrape metrics", DefaultMetricsPath),
		Namespace:   NewField("metrics.namespace", "METRICS_NAMESPACE", "namespace for metrics", DefaultNamespace),
		AppName:     appName,
	}
}

// PaseEnvVars reads the OpenTracing configuration from environment variables
// and sets the values in the configuration
func (c *MetricsConfig) PaseEnvVars() {
	c.MetricsPath.Value = GetEnv(c.MetricsPath.EnVarName, c.MetricsPath.Value)
	c.Namespace.Value = GetEnv(c.Namespace.EnVarName, c.Namespace.Value)
}

func NewMetrics(c *MetricsConfig) *Metrics {
	h := promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	)

	// NOTE: Take care of metrics name
	// https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels
	mtrs := &Metrics{
		Up:         prometheus.NewGauge(prometheus.GaugeOpts{Name: "up", Help: c.AppName + " is up and running."}),
		Http_calls: *prometheus.NewCounterVec(prometheus.CounterOpts{Name: "http_calls", Help: "Number of http calls"}, []string{"path", "code", "method"}),
		handler:    h,
		namespace:  c.Namespace.Value,
	}

	// Register prometheus metrics
	// Global
	prometheus.MustRegister(mtrs.Up)

	prometheus.MustRegister(mtrs.Http_calls)
	return mtrs
}

func (m *Metrics) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /metrics", m.handler)

}
