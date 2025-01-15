package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib" // load the PostgreSQL driver for pgx

	"github.com/p2p-b2b/go-rest-api-service-template/database"
	"github.com/p2p-b2b/go-rest-api-service-template/docs"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/config"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/handler"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/middleware"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/server"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/repository"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/service"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/version"
)

var (
	appName    = "go-rest-api-service-template"
	apiVersion = "v1"
	apiPrefix  = fmt.Sprintf("api/%s", apiVersion)

	LogConfig     = config.NewLogConfig()
	HTTPSrvConfig = config.NewHTTPServerConfig()
	DBConfig      = config.NewDatabaseConfig()
	OTConfig      = config.NewOpenTelemetryConfig(appName, version.Version)

	logHandler        slog.Handler
	logHandlerOptions *slog.HandlerOptions
	logger            *slog.Logger

	showVersion     bool
	showLongVersion bool
	showHelp        bool
	debug           bool
)

func init() {
	// Version flag
	flag.BoolVar(&showVersion, "version", false, "Show the version information")
	flag.BoolVar(&showLongVersion, "version.long", false, "Show the long version information")
	flag.BoolVar(&debug, "debug", false, "Enable debug mode. This is a shorthand for -log.level=debug")
	flag.BoolVar(&showHelp, "help", false, "Show this help message")

	// Log configuration values
	flag.StringVar(&LogConfig.Level.Value, LogConfig.Level.FlagName, config.DefaultLogLevel, LogConfig.Level.FlagDescription)
	flag.StringVar(&LogConfig.Format.Value, LogConfig.Format.FlagName, config.DefaultLogFormat, LogConfig.Format.FlagDescription)
	flag.Var(&LogConfig.Output.Value, LogConfig.Output.FlagName, LogConfig.Output.FlagDescription)

	// HTTP Server configuration values
	flag.StringVar(&HTTPSrvConfig.Address.Value, HTTPSrvConfig.Address.FlagName, config.DefaultHTTPServerAddress, HTTPSrvConfig.Address.FlagDescription)
	flag.IntVar(&HTTPSrvConfig.Port.Value, HTTPSrvConfig.Port.FlagName, config.DefaultHTTPServerPort, HTTPSrvConfig.Port.FlagDescription)
	flag.DurationVar(&HTTPSrvConfig.ShutdownTimeout.Value, HTTPSrvConfig.ShutdownTimeout.FlagName, config.DefaultHTTPServerShutdownTimeout, HTTPSrvConfig.ShutdownTimeout.FlagDescription)
	flag.Var(&HTTPSrvConfig.PrivateKeyFile.Value, HTTPSrvConfig.PrivateKeyFile.FlagName, HTTPSrvConfig.PrivateKeyFile.FlagDescription)
	flag.Var(&HTTPSrvConfig.CertificateFile.Value, HTTPSrvConfig.CertificateFile.FlagName, HTTPSrvConfig.CertificateFile.FlagDescription)
	flag.BoolVar(&HTTPSrvConfig.TLSEnabled.Value, HTTPSrvConfig.TLSEnabled.FlagName, config.DefaultHTTPServerTLSEnabled, HTTPSrvConfig.TLSEnabled.FlagDescription)
	flag.BoolVar(&HTTPSrvConfig.PprofEnabled.Value, HTTPSrvConfig.PprofEnabled.FlagName, config.DefaultHTTPServerPprofEnabled, HTTPSrvConfig.PprofEnabled.FlagDescription)
	flag.BoolVar(&HTTPSrvConfig.CorsEnabled.Value, HTTPSrvConfig.CorsEnabled.FlagName, config.DefaultHTTPServerCorsEnabled, HTTPSrvConfig.CorsEnabled.FlagDescription)
	flag.BoolVar(&HTTPSrvConfig.CorsAllowCredentials.Value, HTTPSrvConfig.CorsAllowCredentials.FlagName, config.DefaultHTTPServerCorsAllowCredentials, HTTPSrvConfig.CorsAllowCredentials.FlagDescription)
	flag.StringVar(&HTTPSrvConfig.CorsAllowedOrigins.Value, HTTPSrvConfig.CorsAllowedOrigins.FlagName, config.DefaultHTTPServerCorsAllowedOrigins, HTTPSrvConfig.CorsAllowedOrigins.FlagDescription)
	flag.StringVar(&HTTPSrvConfig.CorsAllowedMethods.Value, HTTPSrvConfig.CorsAllowedMethods.FlagName, config.DefaultHTTPServerCorsAllowedMethods, HTTPSrvConfig.CorsAllowedMethods.FlagDescription)
	flag.StringVar(&HTTPSrvConfig.CorsAllowedHeaders.Value, HTTPSrvConfig.CorsAllowedHeaders.FlagName, config.DefaultHTTPServerCorsAllowedHeaders, HTTPSrvConfig.CorsAllowedHeaders.FlagDescription)

	// Database configuration values
	flag.StringVar(&DBConfig.Kind.Value, DBConfig.Kind.FlagName, config.DefaultDatabaseKind, DBConfig.Kind.FlagDescription)
	flag.StringVar(&DBConfig.Address.Value, DBConfig.Address.FlagName, config.DefaultDatabaseAddress, DBConfig.Address.FlagDescription)
	flag.IntVar(&DBConfig.Port.Value, DBConfig.Port.FlagName, config.DefaultDatabasePort, DBConfig.Port.FlagDescription)
	flag.StringVar(&DBConfig.Username.Value, DBConfig.Username.FlagName, config.DefaultDatabaseUsername, DBConfig.Username.FlagDescription)
	flag.StringVar(&DBConfig.Password.Value, DBConfig.Password.FlagName, config.DefaultDatabasePassword, DBConfig.Password.FlagDescription)
	flag.StringVar(&DBConfig.Name.Value, DBConfig.Name.FlagName, config.DefaultDatabaseName, DBConfig.Name.FlagDescription)
	flag.StringVar(&DBConfig.SSLMode.Value, DBConfig.SSLMode.FlagName, config.DefaultDatabaseSSLMode, DBConfig.SSLMode.FlagDescription)
	flag.StringVar(&DBConfig.TimeZone.Value, DBConfig.TimeZone.FlagName, config.DefaultDatabaseTimeZone, DBConfig.TimeZone.FlagDescription)
	flag.DurationVar(&DBConfig.MaxPingTimeout.Value, DBConfig.MaxPingTimeout.FlagName, config.DefaultDatabaseMaxPingTimeout, DBConfig.MaxPingTimeout.FlagDescription)
	flag.DurationVar(&DBConfig.MaxQueryTimeout.Value, DBConfig.MaxQueryTimeout.FlagName, config.DefaultDatabaseMaxQueryTimeout, DBConfig.MaxQueryTimeout.FlagDescription)
	flag.DurationVar(&DBConfig.ConnMaxLifetime.Value, DBConfig.ConnMaxLifetime.FlagName, config.DefaultDatabaseConnMaxLifetime, DBConfig.ConnMaxLifetime.FlagDescription)
	flag.IntVar(&DBConfig.MaxIdleConns.Value, DBConfig.MaxIdleConns.FlagName, config.DefaultDatabaseMaxIdleConns, DBConfig.MaxIdleConns.FlagDescription)
	flag.IntVar(&DBConfig.MaxOpenConns.Value, DBConfig.MaxOpenConns.FlagName, config.DefaultDatabaseMaxOpenConns, DBConfig.MaxOpenConns.FlagDescription)
	flag.BoolVar(&DBConfig.MigrationEnable.Value, DBConfig.MigrationEnable.FlagName, config.DefaultDatabaseMigrationEnable, DBConfig.MigrationEnable.FlagDescription)

	// OpenTelemetry configuration values
	flag.StringVar(&OTConfig.TraceEndpoint.Value, OTConfig.TraceEndpoint.FlagName, config.DefaultTraceEndpoint, OTConfig.TraceEndpoint.FlagDescription)
	flag.IntVar(&OTConfig.TracePort.Value, OTConfig.TracePort.FlagName, config.DefaultTracePort, OTConfig.TracePort.FlagDescription)
	flag.StringVar(&OTConfig.TraceExporter.Value, OTConfig.TraceExporter.FlagName, config.DefaultTraceExporter, OTConfig.TraceExporter.FlagDescription)
	flag.DurationVar(&OTConfig.TraceExporterBatchTimeout.Value, OTConfig.TraceExporterBatchTimeout.FlagName, config.DefaultTraceExporterBatchTimeout, OTConfig.TraceExporterBatchTimeout.FlagDescription)
	flag.IntVar(&OTConfig.TraceSampling.Value, OTConfig.TraceSampling.FlagName, config.DefaultTraceSampling, OTConfig.TraceSampling.FlagDescription)
	flag.StringVar(&OTConfig.MetricEndpoint.Value, OTConfig.MetricEndpoint.FlagName, config.DefaultMetricEndpoint, OTConfig.TraceEndpoint.FlagDescription)
	flag.IntVar(&OTConfig.MetricPort.Value, OTConfig.MetricPort.FlagName, config.DefaultMetricPort, OTConfig.MetricPort.FlagDescription)
	flag.StringVar(&OTConfig.MetricExporter.Value, OTConfig.MetricExporter.FlagName, config.DefaultMetricExporter, OTConfig.MetricExporter.FlagDescription)
	flag.DurationVar(&OTConfig.MetricInterval.Value, OTConfig.MetricInterval.FlagName, config.DefaultMetricInterval, OTConfig.MetricInterval.FlagDescription)

	// Parse the command line arguments
	flag.Parse()

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options]\n\nOptions:\n", appName)
		flag.PrintDefaults()
	}

	// implement the version flag
	if showVersion {
		fmt.Printf("%s version: %s\n", appName, version.Version)
		os.Exit(0)
	}

	// implement the long version flag
	if showLongVersion {
		fmt.Printf("%s version: %s,  Git Commit: %s, Build Date: %s, Go Version: %s, OS/Arch: %s/%s\n",
			appName,
			version.Version,
			version.GitCommit,
			version.BuildDate,
			version.GoVersion,
			version.GoVersionOS,
			version.GoVersionArch,
		)
		os.Exit(0)
	}

	// implement the help flag
	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// Set the log level
	if debug {
		LogConfig.Level.Value = "debug"
	}

	switch strings.ToLower(LogConfig.Level.Value) {
	case "debug":
		logHandlerOptions = &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}
	case "info":
		logHandlerOptions = &slog.HandlerOptions{Level: slog.LevelInfo}
	case "warn":
		logHandlerOptions = &slog.HandlerOptions{Level: slog.LevelWarn}
	case "error":
		logHandlerOptions = &slog.HandlerOptions{Level: slog.LevelError, AddSource: true}
	default:
		slog.Error("invalid log level", "level", LogConfig.Level.Value)
	}

	// Set the log format and output
	switch strings.ToLower(LogConfig.Format.Value) {
	case "text":
		logHandler = slog.NewTextHandler(LogConfig.Output.Value.File, logHandlerOptions)
	case "json":
		logHandler = slog.NewJSONHandler(LogConfig.Output.Value.File, logHandlerOptions)
	default:
		slog.Error("invalid log format", "format", LogConfig.Format.Value)
	}

	// Set the default logger
	logger = slog.New(logHandler)
	slog.SetDefault(logger)

	// Get Configuration from Environment Variables
	// and override the values when they are set
	if err := config.SetEnvVarFromFile(); err != nil {
		slog.Error("failed to set environment variables from .env file", "error", err)
		os.Exit(1)
	}

	// Get Configuration from Environment Variables
	// and override the values when they are set
	config.ParseEnvVars(LogConfig, HTTPSrvConfig, DBConfig, OTConfig)

	// Validate the configuration
	if err := config.Validate(LogConfig, HTTPSrvConfig, DBConfig, OTConfig); err != nil {
		slog.Error("error validating configuration", "error", err)
		os.Exit(1)
	}
}

// @tile Golang RESTful API Service Template
// @description This is a service template for building RESTful APIs in Go.
// @description It uses a PostgreSQL database to store user information.
// @description The service provides:
// @description - CRUD operations for users.
// @description - Health and version endpoints.
// @description - Configuration using environment variables or command line arguments.
// @description - Debug mode to enable debug logging.
// @description - TLS enabled to secure the communication.
func main() {
	// Default context
	ctx := context.Background()

	// create OpenTelemetry
	telemetry, err := o11y.New(ctx, OTConfig)
	if err != nil {
		slog.Error("error creating OpenTelemetry", "error", err)
		os.Exit(1)
	}

	if err := telemetry.Start(); err != nil {
		slog.Error("error starting telemetry", "error", err)
		os.Exit(1)
	}

	// Configure server URL information
	serverProtocol := "http"
	if HTTPSrvConfig.TLSEnabled.Value {
		serverProtocol = "https"
	}
	serverURL := fmt.Sprintf("%s://%s:%d/%s", serverProtocol, HTTPSrvConfig.Address.Value, HTTPSrvConfig.Port.Value, apiPrefix)
	statusURL := fmt.Sprintf("%s/status", serverURL)
	serverHost := fmt.Sprintf("%s:%d", HTTPSrvConfig.Address.Value, HTTPSrvConfig.Port.Value)
	swaggerURLIndex := fmt.Sprintf("%s/swagger/index.html", serverURL)
	swaggerURLDocs := fmt.Sprintf("%s/swagger/doc.json", serverURL)

	slog.Info("server endpoints",
		"api", serverURL,
		"status", statusURL,
		"swagger", swaggerURLIndex,
	)

	// Configure Swagger metadata
	docs.SwaggerInfo.Host = serverHost
	docs.SwaggerInfo.BasePath = fmt.Sprintf("/%s", apiPrefix)
	docs.SwaggerInfo.Schemes = []string{serverProtocol}
	docs.SwaggerInfo.Version = version.Version

	// Create PGSQLUserStore
	dbDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		DBConfig.Address.Value,
		DBConfig.Port.Value,
		DBConfig.Username.Value,
		DBConfig.Password.Value,
		DBConfig.Name.Value,
		DBConfig.SSLMode.Value,
		DBConfig.TimeZone.Value,
	)

	db, err := sql.Open(DBConfig.Kind.Value, dbDSN)
	if err != nil {
		slog.Error("database connection error", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	db.SetMaxIdleConns(DBConfig.MaxIdleConns.Value)
	db.SetMaxOpenConns(DBConfig.MaxOpenConns.Value)
	db.SetConnMaxLifetime(DBConfig.ConnMaxLifetime.Value)
	db.SetConnMaxIdleTime(DBConfig.ConnMaxIdleTime.Value)

	slog.Debug("database connection",
		"dsn", dbDSN,
		"kind", DBConfig.Kind.Value,
		"address", DBConfig.Address.Value,
		"port", DBConfig.Port.Value,
		"username", DBConfig.Username.Value,
		"name", DBConfig.Name.Value,
		"ssl_mode", DBConfig.SSLMode.Value,
		"max_idle_conns", DBConfig.MaxIdleConns.Value,
		"max_open_conns", DBConfig.MaxOpenConns.Value,
		"conn_max_lifetime", DBConfig.ConnMaxLifetime.Value,
		"conn_max_idle_time", DBConfig.ConnMaxIdleTime.Value,
	)

	// Test database connection
	ctx, cancel := context.WithTimeout(ctx, DBConfig.MaxPingTimeout.Value)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		slog.Error("database ping error",
			"kind", DBConfig.Kind.Value,
			"address", DBConfig.Address.Value,
			"port", DBConfig.Port.Value,
			"username", DBConfig.Username.Value,
			"ssl_mode", DBConfig.SSLMode.Value,
			"max_idle_conns", DBConfig.MaxIdleConns.Value,
			"max_open_conns", DBConfig.MaxOpenConns.Value,
			"conn_max_lifetime", DBConfig.ConnMaxLifetime.Value,
			"conn_max_idle_time", DBConfig.ConnMaxIdleTime.Value,
			"error", err)
		os.Exit(1)
	}

	// Run the database migrations
	if DBConfig.MigrationEnable.Value {
		slog.Info("running database migrations")
		if err := database.Migrate(ctx, DBConfig.Kind.Value, db); err != nil {
			slog.Error("database migration error", "error", err)
			os.Exit(1)
		}
	}

	// Create a new userRepository
	userRepository, err := repository.NewUserRepository(
		repository.UserRepositoryConfig{
			DB:              db,
			MaxPingTimeout:  DBConfig.MaxPingTimeout.Value,
			MaxQueryTimeout: DBConfig.MaxQueryTimeout.Value,
			OT:              telemetry,
		},
	)
	if err != nil {
		slog.Error("error creating user repository", "error", err)
		os.Exit(1)
	}

	// Create user Service config
	userServiceConf := service.UserServiceConf{
		Repository: userRepository,
		OT:         telemetry,
	}

	// Create user Services
	userService, err := service.NewUserService(userServiceConf)
	if err != nil {
		slog.Error("error creating user service", "error", err)
		os.Exit(1)
	}

	// Create handler config
	userHandlerConf := handler.UserHandlerConf{
		Service: userService,
		OT:      telemetry,
	}

	// Create handlers
	versionHandler := handler.NewVersionHandler()
	userHandler, err := handler.NewUserHandler(userHandlerConf)
	if err != nil {
		slog.Error("error creating user handler", "error", err)
		os.Exit(1)
	}
	swaggerHandler := handler.NewSwaggerHandler(swaggerURLDocs)
	pprofHandler := handler.NewPprofHandler()

	// Create a new ServeMux and register the handlers
	apiRouter := http.NewServeMux()

	swaggerHandler.RegisterRoutes(apiRouter)
	versionHandler.RegisterRoutes(apiRouter)
	userHandler.RegisterRoutes(apiRouter)

	if HTTPSrvConfig.PprofEnabled.Value {
		pprofHandler.RegisterRoutes(apiRouter)
	}

	mdws := []middleware.Middleware{
		middleware.RewriteStandardErrorsAsJSON,
		middleware.Logging,
		middleware.HeaderAPIVersion(apiPrefix),
		middleware.OtelTextMapPropagation,
	}

	if HTTPSrvConfig.CorsEnabled.Value {
		slog.Warn("CORS enabled",
			"allowed_origins", HTTPSrvConfig.CorsAllowedOrigins.Value,
			"allowed_methods", HTTPSrvConfig.CorsAllowedMethods.Value,
			"allowed_headers", HTTPSrvConfig.CorsAllowedHeaders.Value,
			"allow_credentials", HTTPSrvConfig.CorsAllowCredentials.Value,
		)

		corsOpts := middleware.CorsOpts{
			AllowedOrigins:   strings.Split(strings.Trim(HTTPSrvConfig.CorsAllowedOrigins.Value, " "), ","),
			AllowedMethods:   strings.Split(strings.Trim(HTTPSrvConfig.CorsAllowedMethods.Value, " "), ","),
			AllowedHeaders:   strings.Split(strings.Trim(HTTPSrvConfig.CorsAllowedHeaders.Value, " "), ","),
			AllowCredentials: HTTPSrvConfig.CorsAllowCredentials.Value,
		}

		mdws = append(mdws, middleware.Cors(corsOpts))
	}

	// middleware chain
	apiMiddlewares := middleware.Chain(
		mdws...,
	)

	mainRouter := http.NewServeMux()
	mainRouter.Handle(fmt.Sprintf("/%s/", apiPrefix), http.StripPrefix(fmt.Sprintf("/%s", apiPrefix), apiMiddlewares(apiRouter)))

	httpServer := server.NewHTTPServer(
		server.HTTPServerConfig{
			Ctx:         ctx,
			HttpHandler: mainRouter,
			Config:      HTTPSrvConfig,
		},
	)

	// Start the server
	go httpServer.Start()

	// Wait for stopChan to close
	<-httpServer.Wait()

	// Shutdown OpenTelemetry
	slog.Info("shutting down OpenTelemetry")
	telemetry.Shutdown()

	slog.Info("server stopped gracefully")
}
