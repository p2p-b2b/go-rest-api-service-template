package main

import (
	"flag"
	"log/slog"
	"net/http"
	"strings"

	"github.com/wereweare/go-service-template/internal/config"
	"github.com/wereweare/go-service-template/internal/handler"
)

var (
	LogConfig = config.NewLogConfig()
	DBConfig  = config.NewDatabaseConfig()

	logHandler        slog.Handler
	logHandlerOptions *slog.HandlerOptions
)

func init() {
	// Log configuration values
	flag.StringVar(&LogConfig.Level.Value, LogConfig.Level.FlagName, config.DefaultLogLevel, "Log Level [debug, info, warn, error]")
	flag.StringVar(&LogConfig.Format.Value, LogConfig.Format.FlagName, config.DefaultLogFormat, "Log Format [text, json]")
	flag.Var(&LogConfig.Output.Value, LogConfig.Output.FlagName, "Log Output")

	// Initialize the application
	flag.StringVar(&DBConfig.Address.Value, DBConfig.Address.FlagName, config.DefaultDatabaseAddress, "Database IP Address or Hostname")
	flag.IntVar(&DBConfig.Port.Value, DBConfig.Port.FlagName, config.DefaultDatabasePort, "Database Port")
	flag.StringVar(&DBConfig.Username.Value, DBConfig.Username.FlagName, config.DefaultDatabaseUsername, "Database Username")
	flag.StringVar(&DBConfig.Password.Value, DBConfig.Password.FlagName, config.DefaultDatabasePassword, "Database Password")
	flag.StringVar(&DBConfig.Name.Value, DBConfig.Name.FlagName, config.DefaultDatabaseName, "Database Name")
	flag.StringVar(&DBConfig.SSLMode.Value, DBConfig.SSLMode.FlagName, config.DefaultDatabaseSSLMode, "Database SSL Mode")
	flag.DurationVar(&DBConfig.MaxPingTimeout.Value, DBConfig.MaxPingTimeout.FlagName, config.DefaultDatabaseMaxPingTimeout, "Database Max Ping Timeout")
	flag.DurationVar(&DBConfig.MaxQueryTimeout.Value, DBConfig.MaxQueryTimeout.FlagName, config.DefaultDatabaseMaxQueryTimeout, "Database Max Query Timeout")
	flag.DurationVar(&DBConfig.ConnMaxLifetime.Value, DBConfig.ConnMaxLifetime.FlagName, config.DefaultDatabaseConnMaxLifetime, "Database Connection Max Lifetime")
	flag.IntVar(&DBConfig.MaxIdleConns.Value, DBConfig.MaxIdleConns.FlagName, config.DefaultDatabaseMaxIdleConns, "Database Max Idle Connections")
	flag.IntVar(&DBConfig.MaxOpenConns.Value, DBConfig.MaxOpenConns.FlagName, config.DefaultDatabaseMaxOpenConns, "Database Max Open Connections")

	flag.Parse()

	// Get Configuration from Environment Variables
	// and override the values when they are set
	DBConfig.PaseEnvVars()
	LogConfig.ParseEnvVars()

	// Set the log level
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

	slog.Info("starting service...")
	slog.Debug("configuration", "value", DBConfig)

	mux := http.NewServeMux()

	// Add the routes
	mux.HandleFunc("GET /version", handler.GetVersion)

	// Configure the server
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start the server
	if err := server.ListenAndServe(); err != nil {
		slog.Error("server error", "error", err)
	}
}
