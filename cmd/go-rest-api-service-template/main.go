package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

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

	pgOnce sync.Once

	logConfig     = config.NewLogConfig()
	httpSrvConfig = config.NewHTTPServerConfig()
	dbConfig      = config.NewDatabaseConfig()
	otConfig      = config.NewOpenTelemetryConfig(appName, version.Version)

	logHandler        slog.Handler
	logHandlerOptions *slog.HandlerOptions
	logger            *slog.Logger

	showVersion     bool
	showLongVersion bool
	showHelp        bool
	debug           bool
)

func flagsConfig() {
	// Get Configuration from Environment Variables
	// and override the values when they are set
	if err := config.SetEnvVarFromFile(); err != nil {
		slog.Error("failed to set environment variables from .env file", "error", err)
		os.Exit(1)
	}

	// Get Configuration from Environment Variables
	// and override the values when they are set
	config.ParseEnvVars(logConfig, httpSrvConfig, dbConfig, otConfig)

	// Version flag
	flag.BoolVar(&showVersion, "version", false, "Show the version information")
	flag.BoolVar(&showLongVersion, "version.long", false, "Show the long version information")
	flag.BoolVar(&debug, "debug", false, "Enable debug mode. This is a shorthand for -log.level=debug")
	flag.BoolVar(&showHelp, "help", false, "Show this help message")

	// Log configuration values
	flag.StringVar(&logConfig.Level.Value, logConfig.Level.FlagName, config.DefaultLogLevel, logConfig.Level.FlagDescription)
	flag.StringVar(&logConfig.Format.Value, logConfig.Format.FlagName, config.DefaultLogFormat, logConfig.Format.FlagDescription)
	flag.Var(&logConfig.Output.Value, logConfig.Output.FlagName, logConfig.Output.FlagDescription)

	// HTTP Server configuration values
	flag.StringVar(&httpSrvConfig.Address.Value, httpSrvConfig.Address.FlagName, config.DefaultHTTPServerAddress, httpSrvConfig.Address.FlagDescription)
	flag.IntVar(&httpSrvConfig.Port.Value, httpSrvConfig.Port.FlagName, config.DefaultHTTPServerPort, httpSrvConfig.Port.FlagDescription)
	flag.DurationVar(&httpSrvConfig.ShutdownTimeout.Value, httpSrvConfig.ShutdownTimeout.FlagName, config.DefaultHTTPServerShutdownTimeout, httpSrvConfig.ShutdownTimeout.FlagDescription)
	flag.Var(&httpSrvConfig.PrivateKeyFile.Value, httpSrvConfig.PrivateKeyFile.FlagName, httpSrvConfig.PrivateKeyFile.FlagDescription)
	flag.Var(&httpSrvConfig.CertificateFile.Value, httpSrvConfig.CertificateFile.FlagName, httpSrvConfig.CertificateFile.FlagDescription)
	flag.BoolVar(&httpSrvConfig.TLSEnabled.Value, httpSrvConfig.TLSEnabled.FlagName, config.DefaultHTTPServerTLSEnabled, httpSrvConfig.TLSEnabled.FlagDescription)
	flag.BoolVar(&httpSrvConfig.PprofEnabled.Value, httpSrvConfig.PprofEnabled.FlagName, config.DefaultHTTPServerPprofEnabled, httpSrvConfig.PprofEnabled.FlagDescription)
	flag.BoolVar(&httpSrvConfig.CorsEnabled.Value, httpSrvConfig.CorsEnabled.FlagName, config.DefaultHTTPServerCorsEnabled, httpSrvConfig.CorsEnabled.FlagDescription)
	flag.BoolVar(&httpSrvConfig.CorsAllowCredentials.Value, httpSrvConfig.CorsAllowCredentials.FlagName, config.DefaultHTTPServerCorsAllowCredentials, httpSrvConfig.CorsAllowCredentials.FlagDescription)
	flag.StringVar(&httpSrvConfig.CorsAllowedOrigins.Value, httpSrvConfig.CorsAllowedOrigins.FlagName, config.DefaultHTTPServerCorsAllowedOrigins, httpSrvConfig.CorsAllowedOrigins.FlagDescription)
	flag.StringVar(&httpSrvConfig.CorsAllowedMethods.Value, httpSrvConfig.CorsAllowedMethods.FlagName, config.DefaultHTTPServerCorsAllowedMethods, httpSrvConfig.CorsAllowedMethods.FlagDescription)
	flag.StringVar(&httpSrvConfig.CorsAllowedHeaders.Value, httpSrvConfig.CorsAllowedHeaders.FlagName, config.DefaultHTTPServerCorsAllowedHeaders, httpSrvConfig.CorsAllowedHeaders.FlagDescription)

	// Database configuration values
	flag.StringVar(&dbConfig.Kind.Value, dbConfig.Kind.FlagName, config.DefaultDatabaseKind, dbConfig.Kind.FlagDescription)
	flag.StringVar(&dbConfig.Address.Value, dbConfig.Address.FlagName, config.DefaultDatabaseAddress, dbConfig.Address.FlagDescription)
	flag.IntVar(&dbConfig.Port.Value, dbConfig.Port.FlagName, config.DefaultDatabasePort, dbConfig.Port.FlagDescription)
	flag.StringVar(&dbConfig.Username.Value, dbConfig.Username.FlagName, config.DefaultDatabaseUsername, dbConfig.Username.FlagDescription)
	flag.StringVar(&dbConfig.Password.Value, dbConfig.Password.FlagName, config.DefaultDatabasePassword, dbConfig.Password.FlagDescription)
	flag.StringVar(&dbConfig.Name.Value, dbConfig.Name.FlagName, config.DefaultDatabaseName, dbConfig.Name.FlagDescription)
	flag.StringVar(&dbConfig.SSLMode.Value, dbConfig.SSLMode.FlagName, config.DefaultDatabaseSSLMode, dbConfig.SSLMode.FlagDescription)
	flag.StringVar(&dbConfig.TimeZone.Value, dbConfig.TimeZone.FlagName, config.DefaultDatabaseTimeZone, dbConfig.TimeZone.FlagDescription)
	flag.DurationVar(&dbConfig.MaxPingTimeout.Value, dbConfig.MaxPingTimeout.FlagName, config.DefaultDatabaseMaxPingTimeout, dbConfig.MaxPingTimeout.FlagDescription)
	flag.DurationVar(&dbConfig.MaxQueryTimeout.Value, dbConfig.MaxQueryTimeout.FlagName, config.DefaultDatabaseMaxQueryTimeout, dbConfig.MaxQueryTimeout.FlagDescription)
	flag.DurationVar(&dbConfig.ConnMaxLifetime.Value, dbConfig.ConnMaxLifetime.FlagName, config.DefaultDatabaseConnMaxLifetime, dbConfig.ConnMaxLifetime.FlagDescription)
	flag.IntVar(&dbConfig.MaxConns.Value, dbConfig.MaxConns.FlagName, config.DefaultDatabaseMaxConns, dbConfig.MaxConns.FlagDescription)
	flag.IntVar(&dbConfig.MinConns.Value, dbConfig.MinConns.FlagName, config.DefaultDatabaseMinConns, dbConfig.MinConns.FlagDescription)
	flag.BoolVar(&dbConfig.MigrationEnable.Value, dbConfig.MigrationEnable.FlagName, config.DefaultDatabaseMigrationEnable, dbConfig.MigrationEnable.FlagDescription)

	// OpenTelemetry configuration values
	flag.StringVar(&otConfig.TraceEndpoint.Value, otConfig.TraceEndpoint.FlagName, config.DefaultTraceEndpoint, otConfig.TraceEndpoint.FlagDescription)
	flag.IntVar(&otConfig.TracePort.Value, otConfig.TracePort.FlagName, config.DefaultTracePort, otConfig.TracePort.FlagDescription)
	flag.StringVar(&otConfig.TraceExporter.Value, otConfig.TraceExporter.FlagName, config.DefaultTraceExporter, otConfig.TraceExporter.FlagDescription)
	flag.DurationVar(&otConfig.TraceExporterBatchTimeout.Value, otConfig.TraceExporterBatchTimeout.FlagName, config.DefaultTraceExporterBatchTimeout, otConfig.TraceExporterBatchTimeout.FlagDescription)
	flag.IntVar(&otConfig.TraceSampling.Value, otConfig.TraceSampling.FlagName, config.DefaultTraceSampling, otConfig.TraceSampling.FlagDescription)
	flag.StringVar(&otConfig.MetricEndpoint.Value, otConfig.MetricEndpoint.FlagName, config.DefaultMetricEndpoint, otConfig.TraceEndpoint.FlagDescription)
	flag.IntVar(&otConfig.MetricPort.Value, otConfig.MetricPort.FlagName, config.DefaultMetricPort, otConfig.MetricPort.FlagDescription)
	flag.StringVar(&otConfig.MetricExporter.Value, otConfig.MetricExporter.FlagName, config.DefaultMetricExporter, otConfig.MetricExporter.FlagDescription)
	flag.DurationVar(&otConfig.MetricInterval.Value, otConfig.MetricInterval.FlagName, config.DefaultMetricInterval, otConfig.MetricInterval.FlagDescription)

	// Parse the command line arguments
	flag.Parse()

	flag.Usage = func() {
		_, err := fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options]\n\nOptions:\n", appName)
		if err != nil {
			slog.Error("error printing usage", "error", err)
			os.Exit(1)
		}

		flag.PrintDefaults()
	}

	// implement the version flag
	if showVersion {
		_, err := fmt.Printf("%s version: %s\n", appName, version.Version)
		if err != nil {
			slog.Error("error printing version", "error", err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	// implement the long version flag
	if showLongVersion {
		_, err := fmt.Printf("%s version: %s,  Git Commit: %s, Build Date: %s, Go Version: %s, OS/Arch: %s/%s\n",
			appName,
			version.Version,
			version.GitCommit,
			version.BuildDate,
			version.GoVersion,
			version.GoVersionOS,
			version.GoVersionArch,
		)
		if err != nil {
			slog.Error("error printing long version", "error", err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	// implement the help flag
	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// Set the log level
	if debug {
		logConfig.Level.Value = "debug"
	}

	switch strings.ToLower(logConfig.Level.Value) {
	case "debug":
		logHandlerOptions = &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}
	case "info":
		logHandlerOptions = &slog.HandlerOptions{Level: slog.LevelInfo}
	case "warn":
		logHandlerOptions = &slog.HandlerOptions{Level: slog.LevelWarn}
	case "error":
		logHandlerOptions = &slog.HandlerOptions{Level: slog.LevelError, AddSource: true}
	default:
		slog.Error("invalid log level", "level", logConfig.Level.Value)
	}

	// Set the log format and output
	switch strings.ToLower(logConfig.Format.Value) {
	case "text":
		logHandler = slog.NewTextHandler(logConfig.Output.Value.File, logHandlerOptions)
	case "json":
		logHandler = slog.NewJSONHandler(logConfig.Output.Value.File, logHandlerOptions)
	default:
		slog.Error("invalid log format", "format", logConfig.Format.Value)
	}

	// Set the default logger
	logger = slog.New(logHandler)
	slog.SetDefault(logger)

	// Validate the configuration
	if err := config.Validate(logConfig, httpSrvConfig, dbConfig, otConfig); err != nil {
		slog.Error("error validating configuration", "error", err)
		os.Exit(1)
	}
}

// main is the entry point of the application
//
//	@title			Go REST API Service Template
//	@version		v1
//	@contact.name	API Support
//	@contact.url	https://qu3ry.me
//	@contact.email	info@qu3ry.me
//	@description	This is a service template for building RESTful APIs in Go.
//	@description	It uses a PostgreSQL database to store user information.
//	@description	The service provides:
//	@description	- CRUD operations for users.
//	@description	- Health and version endpoints.
//	@description	- Configuration using environment variables or command line arguments.
//	@description	- Debug mode to enable debug logging.
//	@description	- TLS enabled to secure the communication.
func main() {
	flagsConfig()

	// Default context
	ctx := context.Background()

	// create OpenTelemetry
	telemetry, err := o11y.New(ctx, otConfig)
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
	if httpSrvConfig.TLSEnabled.Value {
		serverProtocol = "https"
	}
	serverURL := fmt.Sprintf("%s://%s:%d/%s", serverProtocol, httpSrvConfig.Address.Value, httpSrvConfig.Port.Value, apiPrefix)
	statusURL := fmt.Sprintf("%s/status", serverURL)
	serverHost := fmt.Sprintf("%s:%d", httpSrvConfig.Address.Value, httpSrvConfig.Port.Value)
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
		dbConfig.Address.Value,
		dbConfig.Port.Value,
		dbConfig.Username.Value,
		dbConfig.Password.Value,
		dbConfig.Name.Value,
		dbConfig.SSLMode.Value,
		dbConfig.TimeZone.Value,
	)

	dbCfg, err := pgxpool.ParseConfig(dbDSN)
	if err != nil {
		slog.Error("error parsing pgx pool config", "error", err)
		os.Exit(1)
	}
	dbCfg.MaxConns = int32(dbConfig.MaxConns.Value)
	dbCfg.MinConns = int32(dbConfig.MinConns.Value)
	dbCfg.MaxConnLifetime = dbConfig.ConnMaxLifetime.Value
	dbCfg.MaxConnIdleTime = dbConfig.ConnMaxIdleTime.Value

	// Ensure singleton instance of the pool
	var dbpool *pgxpool.Pool
	pgOnce.Do(func() {
		dbpool, err = pgxpool.NewWithConfig(ctx, dbCfg)
		if err != nil {
			slog.Error("database connection error", "error", err)
			os.Exit(1)
		}
		// defer dbpool.Close()
	})

	slog.Debug("database connection",
		"dsn", dbDSN,
		"kind", dbConfig.Kind.Value,
		"address", dbConfig.Address.Value,
		"port", dbConfig.Port.Value,
		"username", dbConfig.Username.Value,
		"name", dbConfig.Name.Value,
		"ssl_mode", dbConfig.SSLMode.Value,
		"max_conns", dbConfig.MaxConns.Value,
		"min_conns", dbConfig.MinConns.Value,
		"conn_max_lifetime", dbConfig.ConnMaxLifetime.Value,
		"conn_max_idle_time", dbConfig.ConnMaxIdleTime.Value,
	)

	// Test database connection
	dbPingCtx, cancel := context.WithTimeout(ctx, dbConfig.MaxPingTimeout.Value)
	defer cancel()

	if err := dbpool.Ping(dbPingCtx); err != nil {
		slog.Error("database ping error",
			"kind", dbConfig.Kind.Value,
			"address", dbConfig.Address.Value,
			"port", dbConfig.Port.Value,
			"username", dbConfig.Username.Value,
			"ssl_mode", dbConfig.SSLMode.Value,
			"max_idle_conns", dbConfig.MaxConns.Value,
			"max_open_conns", dbConfig.MinConns.Value,
			"conn_max_lifetime", dbConfig.ConnMaxLifetime.Value,
			"conn_max_idle_time", dbConfig.ConnMaxIdleTime.Value,
			"error", err)
		os.Exit(1)
	}

	// Run the database migrations
	if dbConfig.MigrationEnable.Value {
		slog.Info("running database migrations")

		db := stdlib.OpenDBFromPool(dbpool)
		if err := database.Migrate(ctx, "pgx", db); err != nil {
			slog.Error("database migration error", "error", err)
			os.Exit(1)
		}
	}

	// Create a new usersRepository
	healthRepository, err := repository.NewHealthRepository(
		repository.HealthRepositoryConfig{
			DB:             dbpool,
			MaxPingTimeout: dbConfig.MaxPingTimeout.Value,
			OT:             telemetry,
		},
	)
	if err != nil {
		slog.Error("error creating health repository", "error", err)
		os.Exit(1)
	}

	usersRepository, err := repository.NewUsersRepository(
		repository.UsersRepositoryConfig{
			DB:              dbpool,
			MaxPingTimeout:  dbConfig.MaxPingTimeout.Value,
			MaxQueryTimeout: dbConfig.MaxQueryTimeout.Value,
			OT:              telemetry,
		},
	)
	if err != nil {
		slog.Error("error creating user repository", "error", err)
		os.Exit(1)
	}

	// Create user Service config
	healthServiceConf := service.HealthServiceConf{
		Repository: healthRepository,
		OT:         telemetry,
	}

	healthService, err := service.NewHealthService(healthServiceConf)
	if err != nil {
		slog.Error("error creating health service", "error", err)
		os.Exit(1)
	}

	usersServiceConf := service.UsersServiceConf{
		Repository: usersRepository,
		OT:         telemetry,
	}

	// Create user Services
	usersService, err := service.NewUsersService(usersServiceConf)
	if err != nil {
		slog.Error("error creating user service", "error", err)
		os.Exit(1)
	}

	// Create handlers
	healthHandlerConf := handler.HealthHandlerConf{
		Service: healthService,
		OT:      telemetry,
	}

	healthHandler, err := handler.NewHealthHandler(healthHandlerConf)
	if err != nil {
		slog.Error("could not create health handler", "error", err)
		os.Exit(1)
	}

	usersHandlerConf := handler.UsersHandlerConf{
		Service: usersService,
		OT:      telemetry,
	}

	usersHandler, err := handler.NewUsersHandler(usersHandlerConf)
	if err != nil {
		slog.Error("error creating user handler", "error", err)
		os.Exit(1)
	}

	versionHandler := handler.NewVersionHandler()
	swaggerHandler := handler.NewSwaggerHandler(swaggerURLDocs)
	pprofHandler := handler.NewPprofHandler()

	// Create a new ServeMux and register the handlers
	apiRouter := http.NewServeMux()

	swaggerHandler.RegisterRoutes(apiRouter)
	versionHandler.RegisterRoutes(apiRouter)
	usersHandler.RegisterRoutes(apiRouter)
	healthHandler.RegisterRoutes(apiRouter)

	if httpSrvConfig.PprofEnabled.Value {
		pprofHandler.RegisterRoutes(apiRouter)
	}

	apiCommonMdws := []middleware.Middleware{
		middleware.RewriteStandardErrorsAsJSON,
		middleware.Logging,
		middleware.HeaderAPIVersion(apiPrefix),
		middleware.OtelTextMapPropagation,
	}

	if httpSrvConfig.CorsEnabled.Value {
		slog.Warn("CORS enabled",
			"allowed_origins", httpSrvConfig.CorsAllowedOrigins.Value,
			"allowed_methods", httpSrvConfig.CorsAllowedMethods.Value,
			"allowed_headers", httpSrvConfig.CorsAllowedHeaders.Value,
			"allow_credentials", httpSrvConfig.CorsAllowCredentials.Value,
		)

		corsOpts := middleware.CorsOpts{
			AllowedOrigins:   strings.Split(strings.Trim(httpSrvConfig.CorsAllowedOrigins.Value, " "), ","),
			AllowedMethods:   strings.Split(strings.Trim(httpSrvConfig.CorsAllowedMethods.Value, " "), ","),
			AllowedHeaders:   strings.Split(strings.Trim(httpSrvConfig.CorsAllowedHeaders.Value, " "), ","),
			AllowCredentials: httpSrvConfig.CorsAllowCredentials.Value,
		}

		apiCommonMdws = append(apiCommonMdws, middleware.Cors(corsOpts))
	}

	// middleware chain
	apiCommonMiddlewares := middleware.Chain(
		apiCommonMdws...,
	)

	mainRouter := http.NewServeMux()
	mainRouter.Handle(fmt.Sprintf("/%s/", apiPrefix), http.StripPrefix(fmt.Sprintf("/%s", apiPrefix), apiCommonMiddlewares(apiRouter)))

	httpServer := server.NewHTTPServer(
		server.HTTPServerConfig{
			Ctx:         ctx,
			HttpHandler: mainRouter,
			Config:      httpSrvConfig,
		},
	)

	// Start the server
	go httpServer.Start()

	// Wait for stopChan to close
	<-httpServer.Wait()

	// close db connection
	slog.Info("closing database connection")
	dbpool.Close()

	// Shutdown OpenTelemetry
	slog.Info("shutting down OpenTelemetry")
	telemetry.Shutdown()

	slog.Info("server stopped gracefully")
}
