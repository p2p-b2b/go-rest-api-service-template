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

	_ "github.com/lib/pq" // load the PostgreSQL driver for database/sql

	// _ "github.com/go-sql-driver/mysql" // load the MySQL driver for database/sql

	"github.com/p2p-b2b/go-rest-api-service-template/database"
	"github.com/p2p-b2b/go-rest-api-service-template/docs"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/config"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/handler"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/opentracing"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/repository"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/server"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/service"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/version"
)

var (
	appName = "go-rest-api-service-template"

	LogConfig = config.NewLogConfig()
	SrvConfig = config.NewServerConfig()
	DBConfig  = config.NewDatabaseConfig()
	OTConfig  = config.NewOpenTracingConfig(appName)

	logHandler        slog.Handler
	logHandlerOptions *slog.HandlerOptions
	logger            *slog.Logger
)

func init() {
	// Log configuration values
	flag.StringVar(&LogConfig.Level.Value, LogConfig.Level.FlagName, config.DefaultLogLevel, LogConfig.Level.FlagDescription)
	flag.StringVar(&LogConfig.Format.Value, LogConfig.Format.FlagName, config.DefaultLogFormat, LogConfig.Format.FlagDescription)
	flag.Var(&LogConfig.Output.Value, LogConfig.Output.FlagName, LogConfig.Output.FlagDescription)

	// Version flag
	flag.Bool("version", false, "Show the version information")
	flag.Bool("version.long", false, "Show the long version information")
	flag.Bool("debug", false, "Enable debug mode. This is a shorthand for -log.level=debug")

	// Server configuration values
	flag.StringVar(&SrvConfig.Address.Value, SrvConfig.Address.FlagName, config.DefaultServerAddress, SrvConfig.Address.FlagDescription)
	flag.IntVar(&SrvConfig.Port.Value, SrvConfig.Port.FlagName, config.DefaultServerPort, SrvConfig.Port.FlagDescription)
	flag.DurationVar(&SrvConfig.ShutdownTimeout.Value, SrvConfig.ShutdownTimeout.FlagName, config.DefaultShutdownTimeout, SrvConfig.ShutdownTimeout.FlagDescription)
	flag.Var(&SrvConfig.PrivateKeyFile.Value, SrvConfig.PrivateKeyFile.FlagName, SrvConfig.PrivateKeyFile.FlagDescription)
	flag.Var(&SrvConfig.CertificateFile.Value, SrvConfig.CertificateFile.FlagName, SrvConfig.CertificateFile.FlagDescription)
	flag.BoolVar(&SrvConfig.TLSEnabled.Value, SrvConfig.TLSEnabled.FlagName, config.DefaultServerTLSEnabled, SrvConfig.TLSEnabled.FlagDescription)
	flag.BoolVar(&SrvConfig.PprofEnabled.Value, SrvConfig.PprofEnabled.FlagName, config.DefaultServerPprofEnabled, SrvConfig.PprofEnabled.FlagDescription)

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

	//Opentrace configuration values
	flag.StringVar(&OTConfig.OTLPEndpoint.Value, OTConfig.OTLPEndpoint.FlagName, config.DefaultOTLPEndpoint, OTConfig.OTLPEndpoint.FlagDescription)
	flag.IntVar(&OTConfig.OTLPPort.Value, OTConfig.OTLPPort.FlagName, config.DefaultOTLPPort, OTConfig.OTLPPort.FlagDescription)

	// Parse the command line arguments
	flag.Bool("help", false, "Show this help message")
	flag.Parse()

	// implement the version flag
	if flag.Lookup("version").Value.(flag.Getter).Get().(bool) {
		fmt.Printf("%s version: %s\n", appName, version.Version)
		os.Exit(0)
	}

	// implement the long version flag
	if flag.Lookup("version.long").Value.(flag.Getter).Get().(bool) {
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
	if flag.Lookup("help").Value.(flag.Getter).Get().(bool) {
		flag.Usage()
		os.Exit(0)
	}

	// validate the database kind
	if DBConfig.Kind.Value != "postgres" && DBConfig.Kind.Value != "mysql" {
		slog.Error("Invalid database kind. Use --help to get more info", "kind", DBConfig.Kind.Value)
		os.Exit(1)
	}

	// Get Configuration from Environment Variables
	// and override the values when they are set
	DBConfig.PaseEnvVars()
	LogConfig.ParseEnvVars()
	OTConfig.PaseEnvVars()

	// Set the log level
	if flag.Lookup("debug").Value.(flag.Getter).Get().(bool) {
		LogConfig.Level.Value = "debug"
	}

	switch strings.ToLower(LogConfig.Level.Value) {
	case "debug":
		logHandlerOptions = &slog.HandlerOptions{Level: slog.LevelDebug}
	case "info":
		logHandlerOptions = &slog.HandlerOptions{Level: slog.LevelInfo}
	case "warn":
		logHandlerOptions = &slog.HandlerOptions{Level: slog.LevelWarn}
	case "error":
		logHandlerOptions = &slog.HandlerOptions{Level: slog.LevelError}
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
	// Set the default logger
	logger = slog.New(logHandler)
	slog.SetDefault(logger)

	// Default context
	ctx := context.Background()

	// Configure server URL information
	serverProtocol := "http"
	if SrvConfig.TLSEnabled.Value {
		serverProtocol = "https"
	}
	serverURL := fmt.Sprintf("%s://%s:%d", serverProtocol, SrvConfig.Address.Value, SrvConfig.Port.Value)
	statusURL := fmt.Sprintf("%s/status", serverURL)
	serverHost := fmt.Sprintf("%s:%d", SrvConfig.Address.Value, SrvConfig.Port.Value)
	swaggerURLIndex := fmt.Sprintf("%s/swagger/index.html", serverURL)
	swaggerURLDocs := fmt.Sprintf("%s/swagger/doc.json", serverURL)

	slog.Info("server endpoints",
		"url", serverURL,
		"status", statusURL,
		"swagger", swaggerURLIndex,
	)

	// Configure Swagger metadata
	docs.SwaggerInfo.Host = serverHost
	docs.SwaggerInfo.BasePath = "/"
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
	)
	slog.Debug("database configuration",
		"max_idle_conns", DBConfig.MaxIdleConns.Value,
		"max_open_conns", DBConfig.MaxOpenConns.Value,
		"conn_max_lifetime", DBConfig.ConnMaxLifetime.Value,
		"conn_max_idle_time", DBConfig.ConnMaxIdleTime.Value,
	)

	// Create a new userRepository
	userRepository := repository.NewPGSQLUserRepository(
		repository.PGSQLUserRepositoryConfig{
			DB:              db,
			MaxPingTimeout:  DBConfig.MaxPingTimeout.Value,
			MaxQueryTimeout: DBConfig.MaxQueryTimeout.Value,
		},
	)

	// Test database connection
	ctx, cancel := context.WithTimeout(ctx, DBConfig.MaxPingTimeout.Value)
	defer cancel()

	slog.Info("testing database connection...", "dsn", dbDSN)
	if err := userRepository.PingContext(ctx); err != nil {
		slog.Error("database ping error", "error", err)
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

	//create OpenTrace
	otracing := opentracing.NewOpentracing(OTConfig)
	otracing.SetContext(ctx)

	//Start tracing
	otracing.SetupOTelSDK()

	//Set tracer
	tracer := otracing.GetTracerProvider().Tracer(appName)
	//Create user Service config
	userServiceConf := service.UserConf{
		Repository: userRepository,
		Ot:         tracer,
	}

	// Create user Services
	userService := service.NewUserService(userServiceConf)

	//Create handler config
	userHandlerConf := handler.UserHandlerConf{
		Service: userService,
		Ot:      tracer,
	}

	// Create handlers
	versionHandler := handler.NewVersionHandler()
	healthHandler := handler.NewHealthHandler(userService)
	userHandler := handler.NewUserHandler(userHandlerConf)
	swaggerHandler := handler.NewSwaggerHandler(swaggerURLDocs)
	pprofHandler := handler.NewPprofHandler()

	// Create a new ServeMux and register the handlers
	mux := http.NewServeMux()
	swaggerHandler.RegisterRoutes(mux)
	healthHandler.RegisterRoutes(mux)
	versionHandler.RegisterRoutes(mux)
	userHandler.RegisterRoutes(mux)

	if SrvConfig.PprofEnabled.Value {
		pprofHandler.RegisterRoutes(mux)
	}

	httpServer := server.NewHttpServer(
		server.ServerConfig{
			Ctx:         ctx,
			HttpHandler: mux,
			Config:      SrvConfig,
		})

	// Start the server
	go httpServer.Start()

	// Wait for stopChan to close
	<-httpServer.Wait()
	slog.Info("server stopped gracefully")
}
