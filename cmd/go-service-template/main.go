package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	_ "github.com/go-sql-driver/mysql" // load the MySQL driver for database/sql
	_ "github.com/lib/pq"              // load the PostgreSQL driver for database/sql

	"github.com/p2p-b2b/go-service-template/internal/config"
	"github.com/p2p-b2b/go-service-template/internal/handler"
	"github.com/p2p-b2b/go-service-template/internal/store"
	"github.com/p2p-b2b/go-service-template/internal/version"
)

var (
	LogConfig = config.NewLogConfig()
	SrvConfig = config.NewServerConfig()
	DBConfig  = config.NewDatabaseConfig()

	appName = "go-service-template"

	logHandler        slog.Handler
	logHandlerOptions *slog.HandlerOptions
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
		fmt.Printf("%s version: %s,  Git Commit: %s, Build Date: %s, Go Version: %s, OS/Arch: %s/%s\n", appName, version.Version, version.GitCommit, version.BuildDate, version.GoVersion, version.GoVersionOS, version.GoVersionArch)
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
		slog.Error("Invalid log level", "level", LogConfig.Level.Value)
	}

	// Set the log format and output
	switch strings.ToLower(LogConfig.Format.Value) {
	case "text":
		logHandler = slog.NewTextHandler(LogConfig.Output.Value.File, logHandlerOptions)
	case "json":
		logHandler = slog.NewJSONHandler(LogConfig.Output.Value.File, logHandlerOptions)
	default:
		slog.Error("Invalid log format", "format", LogConfig.Format.Value)
	}
}

func main() {
	// Set the default logger
	slog.SetDefault(slog.New(logHandler))

	slog.Debug("configuration", "database", DBConfig)
	slog.Debug("configuration", "log", LogConfig)

	// Create a new ServeMux
	mux := http.NewServeMux()

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

	// Create a new PGSQLUserStore
	pgsqlUserStore := store.NewPGSQLUserStore(
		store.PGSQLUserStoreConfig{
			DB:              db,
			MaxPingTimeout:  DBConfig.MaxPingTimeout.Value,
			MaxQueryTimeout: DBConfig.MaxQueryTimeout.Value,
		},
	)

	slog.Debug("database connection",
		"dsn", dbDSN,
		"kind", DBConfig.Kind.Value,
		"address", DBConfig.Address.Value,
		"port", DBConfig.Port.Value,
		"username", DBConfig.Username.Value,
		"password", DBConfig.Password.Value,
		"name", DBConfig.Name.Value,
		"ssl_mode", DBConfig.SSLMode.Value,
	)
	slog.Debug("database configuration",
		"max_idle_conns", DBConfig.MaxIdleConns.Value,
		"max_open_conns", DBConfig.MaxOpenConns.Value,
		"conn_max_lifetime", DBConfig.ConnMaxLifetime.Value,
		"conn_max_idle_time", DBConfig.ConnMaxIdleTime.Value,
	)

	// Ping the database to check the connection
	if err := pgsqlUserStore.Ping(context.Background()); err != nil {
		slog.Error("database ping error", "error", err)
		os.Exit(1)
	}

	// Create handlers
	versionHandler := &handler.VersionHandler{}
	mux.HandleFunc("GET /version", versionHandler.Get)

	// Create a new RepositoryHandler
	repositoryHandler := &handler.RepositoryHandler{
		Repository: pgsqlUserStore,
	}
	mux.HandleFunc("GET /users/{id}", repositoryHandler.GetUserByID)
	mux.HandleFunc("POST /users", repositoryHandler.CreateUser)
	mux.HandleFunc("PUT /users/{id}", repositoryHandler.UpdateUser)
	mux.HandleFunc("DELETE /users/{id}", repositoryHandler.DeleteUser)
	mux.HandleFunc("GET /users", repositoryHandler.ListUsers)

	// Configure the server
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", SrvConfig.Address.Value, SrvConfig.Port.Value),
		Handler: mux,
	}

	// Configure the TLS
	if SrvConfig.TLSEnabled.Value {
		slog.Info("configuring tls")
		if _, err := os.Stat(SrvConfig.CertificateFile.Value.Name()); os.IsNotExist(err) {
			slog.Error("tls.crt file not found")
			os.Exit(1)
		}

		if _, err := os.Stat(SrvConfig.PrivateKeyFile.Value.Name()); os.IsNotExist(err) {
			slog.Error("tls.key file not found")
			os.Exit(1)
		}

		tlsCfg := &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		}
		server.TLSConfig = tlsCfg
		server.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)
	}

	// Wait for a signal to shutdown
	osSigChan := make(chan os.Signal, 1)
	signal.Notify(osSigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	stopChan := make(chan struct{})

	ctx, cancel := context.WithTimeout(context.Background(), SrvConfig.ShutdownTimeout.Value)
	defer cancel()

	// Handle signals
	go func() {
		slog.Info("waiting for os signals...")
		for {
			select {
			case sig := <-osSigChan:
				slog.Debug("received signal", "signal", sig)

				// Handle the signal to shutdown the server or reload
				switch sig {
				case os.Interrupt, syscall.SIGINT, syscall.SIGTERM:
					slog.Warn("shutting down server...")
					if err := server.Shutdown(ctx); err != nil {
						slog.Error("server shutdown with error", "error", err)
						os.Exit(1)
					}
					close(stopChan)
					return
				case syscall.SIGHUP:
					slog.Warn("reloading server...")
					// Reload the server
					// This is where you would reload the server
					return
				default:
					slog.Warn("unknown signal", "signal", sig)
					return
				}

			case <-stopChan:
				return
			}
		}
	}()

	// Start the server
	go func() {
		slog.Info("starting server",
			"address", SrvConfig.Address.Value,
			"port", SrvConfig.Port.Value,
			"url", fmt.Sprintf("http://%s:%d", SrvConfig.Address.Value, SrvConfig.Port.Value),
		)

		// Check if the port is 443 and start the server with TLS
		if SrvConfig.TLSEnabled.Value {
			slog.Info("server using tls")
			if err := server.ListenAndServeTLS(
				SrvConfig.CertificateFile.Value.Name(),
				SrvConfig.PrivateKeyFile.Value.Name(),
			); !errors.Is(err, http.ErrServerClosed) {
				slog.Error("server error", "error", err)
				os.Exit(1)
			}
		} else {
			slog.Info("server using http")
			if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
				slog.Error("server error", "error", err)
				os.Exit(1)
			}
		}
	}()

	// Wait for stopChan to close
	<-stopChan
	slog.Info("server stopped gracefully")
}
