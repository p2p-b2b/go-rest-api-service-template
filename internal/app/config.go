package app

import (
	"flag"
	"fmt"
	"log/slog"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/config"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/version"
)

// Configs contains all application configurations
type Configs struct {
	Log        *config.LogConfig
	HTTPServer *config.HTTPServerConfig
	Database   *config.DatabaseConfig
	Telemetry  *config.OpenTelemetryConfig

	ShowVersion     bool
	ShowLongVersion bool
	ShowHelp        bool
	Debug           bool
}

// LoadConfigs loads all configuration from flags and environment variables
func LoadConfigs() (*Configs, error) {
	configs := &Configs{
		Log:        config.NewLogConfig(),
		HTTPServer: config.NewHTTPServerConfig(),
		Database:   config.NewDatabaseConfig(),
		Telemetry:  config.NewOpenTelemetryConfig(appName, version.Version),
	}

	// Register flags
	setupFlags(configs)

	// Parse the command line arguments
	flag.Parse()

	// Handle special flags
	if err := handleSpecialFlags(configs); err != nil {
		return nil, err
	}

	// Load environment variables
	if err := config.SetEnvVarFromFile(); err != nil {
		slog.Error("failed to set environment variables from .env file", "error", err)
		return nil, err
	}

	config.ParseEnvVars(
		configs.Log,
		configs.HTTPServer,
		configs.Database,
		configs.Telemetry,
	)

	// Validate configuration
	if err := config.Validate(
		configs.Log,
		configs.HTTPServer,
		configs.Database,
		configs.Telemetry,
	); err != nil {
		return nil, fmt.Errorf("error validating configuration: %w", err)
	}

	// Setup logger based on configuration
	setupLogger(configs.Log)

	return configs, nil
}

// setupFlags configures command line flags for all application configurations
func setupFlags(configs *Configs) {
	// Version flag
	flag.BoolVar(&configs.ShowVersion, "version", false, "Show the version information")
	flag.BoolVar(&configs.ShowLongVersion, "version.long", false, "Show the long version information")
	flag.BoolVar(&configs.Debug, "debug", false, "Enable debug mode. This is a shorthand for -log.level=debug")
	flag.BoolVar(&configs.ShowHelp, "help", false, "Show this help message")

	// Log configuration values
	flag.StringVar(&configs.Log.Level.Value, configs.Log.Level.FlagName, config.DefaultLogLevel, configs.Log.Level.FlagDescription)
	flag.StringVar(&configs.Log.Format.Value, configs.Log.Format.FlagName, config.DefaultLogFormat, configs.Log.Format.FlagDescription)
	flag.Var(&configs.Log.Output.Value, configs.Log.Output.FlagName, configs.Log.Output.FlagDescription)

	// HTTP Server configuration values
	flag.StringVar(&configs.HTTPServer.Address.Value, configs.HTTPServer.Address.FlagName, config.DefaultHTTPServerAddress, configs.HTTPServer.Address.FlagDescription)
	flag.IntVar(&configs.HTTPServer.Port.Value, configs.HTTPServer.Port.FlagName, config.DefaultHTTPServerPort, configs.HTTPServer.Port.FlagDescription)
	flag.DurationVar(&configs.HTTPServer.ShutdownTimeout.Value, configs.HTTPServer.ShutdownTimeout.FlagName, config.DefaultHTTPServerShutdownTimeout, configs.HTTPServer.ShutdownTimeout.FlagDescription)
	flag.Var(&configs.HTTPServer.PrivateKeyFile.Value, configs.HTTPServer.PrivateKeyFile.FlagName, configs.HTTPServer.PrivateKeyFile.FlagDescription)
	flag.Var(&configs.HTTPServer.CertificateFile.Value, configs.HTTPServer.CertificateFile.FlagName, configs.HTTPServer.CertificateFile.FlagDescription)
	flag.BoolVar(&configs.HTTPServer.TLSEnabled.Value, configs.HTTPServer.TLSEnabled.FlagName, config.DefaultHTTPServerTLSEnabled, configs.HTTPServer.TLSEnabled.FlagDescription)
	flag.StringVar(&configs.HTTPServer.PprofAddress.Value, configs.HTTPServer.PprofAddress.FlagName, config.DefaultHTTPServerPprofAddress, configs.HTTPServer.PprofAddress.FlagDescription)
	flag.IntVar(&configs.HTTPServer.PprofPort.Value, configs.HTTPServer.PprofPort.FlagName, config.DefaultHTTPServerPprofPort, configs.HTTPServer.PprofPort.FlagDescription)
	flag.BoolVar(&configs.HTTPServer.PprofEnabled.Value, configs.HTTPServer.PprofEnabled.FlagName, config.DefaultHTTPServerPprofEnabled, configs.HTTPServer.PprofEnabled.FlagDescription)
	flag.BoolVar(&configs.HTTPServer.CorsEnabled.Value, configs.HTTPServer.CorsEnabled.FlagName, config.DefaultHTTPServerCorsEnabled, configs.HTTPServer.CorsEnabled.FlagDescription)
	flag.BoolVar(&configs.HTTPServer.CorsAllowCredentials.Value, configs.HTTPServer.CorsAllowCredentials.FlagName, config.DefaultHTTPServerCorsAllowCredentials, configs.HTTPServer.CorsAllowCredentials.FlagDescription)
	flag.StringVar(&configs.HTTPServer.CorsAllowedOrigins.Value, configs.HTTPServer.CorsAllowedOrigins.FlagName, config.DefaultHTTPServerCorsAllowedOrigins, configs.HTTPServer.CorsAllowedOrigins.FlagDescription)
	flag.StringVar(&configs.HTTPServer.CorsAllowedMethods.Value, configs.HTTPServer.CorsAllowedMethods.FlagName, config.DefaultHTTPServerCorsAllowedMethods, configs.HTTPServer.CorsAllowedMethods.FlagDescription)
	flag.StringVar(&configs.HTTPServer.CorsAllowedHeaders.Value, configs.HTTPServer.CorsAllowedHeaders.FlagName, config.DefaultHTTPServerCorsAllowedHeaders, configs.HTTPServer.CorsAllowedHeaders.FlagDescription)

	// Database configuration values
	flag.StringVar(&configs.Database.Kind.Value, configs.Database.Kind.FlagName, config.DefaultDatabaseKind, configs.Database.Kind.FlagDescription)
	flag.StringVar(&configs.Database.Address.Value, configs.Database.Address.FlagName, config.DefaultDatabaseAddress, configs.Database.Address.FlagDescription)
	flag.IntVar(&configs.Database.Port.Value, configs.Database.Port.FlagName, config.DefaultDatabasePort, configs.Database.Port.FlagDescription)
	flag.StringVar(&configs.Database.Username.Value, configs.Database.Username.FlagName, config.DefaultDatabaseUsername, configs.Database.Username.FlagDescription)
	flag.StringVar(&configs.Database.Password.Value, configs.Database.Password.FlagName, config.DefaultDatabasePassword, configs.Database.Password.FlagDescription)
	flag.StringVar(&configs.Database.Name.Value, configs.Database.Name.FlagName, config.DefaultDatabaseName, configs.Database.Name.FlagDescription)
	flag.StringVar(&configs.Database.SSLMode.Value, configs.Database.SSLMode.FlagName, config.DefaultDatabaseSSLMode, configs.Database.SSLMode.FlagDescription)
	flag.StringVar(&configs.Database.TimeZone.Value, configs.Database.TimeZone.FlagName, config.DefaultDatabaseTimeZone, configs.Database.TimeZone.FlagDescription)
	flag.DurationVar(&configs.Database.MaxPingTimeout.Value, configs.Database.MaxPingTimeout.FlagName, config.DefaultDatabaseMaxPingTimeout, configs.Database.MaxPingTimeout.FlagDescription)
	flag.DurationVar(&configs.Database.MaxQueryTimeout.Value, configs.Database.MaxQueryTimeout.FlagName, config.DefaultDatabaseMaxQueryTimeout, configs.Database.MaxQueryTimeout.FlagDescription)
	flag.DurationVar(&configs.Database.ConnMaxLifetime.Value, configs.Database.ConnMaxLifetime.FlagName, config.DefaultDatabaseConnMaxLifetime, configs.Database.ConnMaxLifetime.FlagDescription)
	flag.IntVar(&configs.Database.MaxConns.Value, configs.Database.MaxConns.FlagName, config.DefaultDatabaseMaxConns, configs.Database.MaxConns.FlagDescription)
	flag.IntVar(&configs.Database.MinConns.Value, configs.Database.MinConns.FlagName, config.DefaultDatabaseMinConns, configs.Database.MinConns.FlagDescription)
	flag.BoolVar(&configs.Database.MigrationEnable.Value, configs.Database.MigrationEnable.FlagName, config.DefaultDatabaseMigrationEnable, configs.Database.MigrationEnable.FlagDescription)

	// OpenTelemetry configuration values
	flag.StringVar(&configs.Telemetry.TraceEndpoint.Value, configs.Telemetry.TraceEndpoint.FlagName, config.DefaultTraceEndpoint, configs.Telemetry.TraceEndpoint.FlagDescription)
	flag.IntVar(&configs.Telemetry.TracePort.Value, configs.Telemetry.TracePort.FlagName, config.DefaultTracePort, configs.Telemetry.TracePort.FlagDescription)
	flag.StringVar(&configs.Telemetry.TraceExporter.Value, configs.Telemetry.TraceExporter.FlagName, config.DefaultTraceExporter, configs.Telemetry.TraceExporter.FlagDescription)
	flag.DurationVar(&configs.Telemetry.TraceExporterBatchTimeout.Value, configs.Telemetry.TraceExporterBatchTimeout.FlagName, config.DefaultTraceExporterBatchTimeout, configs.Telemetry.TraceExporterBatchTimeout.FlagDescription)
	flag.IntVar(&configs.Telemetry.TraceSampling.Value, configs.Telemetry.TraceSampling.FlagName, config.DefaultTraceSampling, configs.Telemetry.TraceSampling.FlagDescription)
	flag.StringVar(&configs.Telemetry.MetricEndpoint.Value, configs.Telemetry.MetricEndpoint.FlagName, config.DefaultMetricEndpoint, configs.Telemetry.MetricEndpoint.FlagDescription)
	flag.IntVar(&configs.Telemetry.MetricPort.Value, configs.Telemetry.MetricPort.FlagName, config.DefaultMetricPort, configs.Telemetry.MetricPort.FlagDescription)
	flag.StringVar(&configs.Telemetry.MetricExporter.Value, configs.Telemetry.MetricExporter.FlagName, config.DefaultMetricExporter, configs.Telemetry.MetricExporter.FlagDescription)
	flag.DurationVar(&configs.Telemetry.MetricInterval.Value, configs.Telemetry.MetricInterval.FlagName, config.DefaultMetricInterval, configs.Telemetry.MetricInterval.FlagDescription)
}

// handleSpecialFlags handles flags that control application execution flow
// such as version display, help display, etc.
func handleSpecialFlags(configs *Configs) error {
	// Handle version flag
	if configs.ShowVersion {
		fmt.Println(version.Version)
		return fmt.Errorf("version displayed")
	}

	// Handle long version flag
	if configs.ShowLongVersion {
		fmt.Printf("Version: %s\n", version.Version)
		fmt.Printf("Build Date: %s\n", version.BuildDate)
		fmt.Printf("Git Commit: %s\n", version.GitCommit)
		fmt.Printf("Git Branch: %s\n", version.GitBranch)
		fmt.Printf("Go Version: %s\n", version.GoVersion)
		fmt.Printf("OS/Arch: %s/%s\n", version.GoVersionOS, version.GoVersionArch)
		return fmt.Errorf("version displayed")
	}

	// Handle help flag
	if configs.ShowHelp {
		flag.Usage()
		return fmt.Errorf("help displayed")
	}

	// Handle debug flag - set log level to debug if enabled
	if configs.Debug {
		configs.Log.Level.Value = "debug"
	}

	return nil
}

// setupLogger configures the global logger based on the given LogConfig
func setupLogger(logConfig *config.LogConfig) {
	var logLevel slog.Level
	switch logConfig.Level.Value {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	// Create logger options
	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: logConfig.AddSource.Value,
	}

	// Create handler based on format
	var handler slog.Handler
	switch logConfig.Format.Value {
	case "json":
		handler = slog.NewJSONHandler(logConfig.Output.Value, opts)
	case "text":
		handler = slog.NewTextHandler(logConfig.Output.Value, opts)
	default:
		handler = slog.NewTextHandler(logConfig.Output.Value, opts)
	}

	// Set the default logger
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
