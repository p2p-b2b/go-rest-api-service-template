package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/wereweare/go-service-template/internal/config"
	"github.com/wereweare/go-service-template/internal/handler"
)

var (
	LogConfig = config.NewLogConfig()
	SrvConfig = config.NewServerConfig()
	DBConfig  = config.NewDatabaseConfig()

	logHandler        slog.Handler
	logHandlerOptions *slog.HandlerOptions
)

func init() {
	// Log configuration values
	flag.StringVar(&LogConfig.Level.Value, LogConfig.Level.FlagName, config.DefaultLogLevel, LogConfig.Level.FlagDescription)
	flag.StringVar(&LogConfig.Format.Value, LogConfig.Format.FlagName, config.DefaultLogFormat, LogConfig.Format.FlagDescription)
	flag.Var(&LogConfig.Output.Value, LogConfig.Output.FlagName, LogConfig.Output.FlagDescription)

	// Server configuration values
	flag.StringVar(&SrvConfig.Address.Value, SrvConfig.Address.FlagName, config.DefaultServerAddress, SrvConfig.Address.FlagDescription)
	flag.IntVar(&SrvConfig.Port.Value, SrvConfig.Port.FlagName, config.DefaultServerPort, SrvConfig.Port.FlagDescription)
	flag.DurationVar(&SrvConfig.ShutdownTimeout.Value, SrvConfig.ShutdownTimeout.FlagName, config.DefaultShutdownTimeout, SrvConfig.ShutdownTimeout.FlagDescription)
	flag.Var(&SrvConfig.PrivateKeyFile.Value, SrvConfig.PrivateKeyFile.FlagName, SrvConfig.PrivateKeyFile.FlagDescription)
	flag.Var(&SrvConfig.CertificateFile.Value, SrvConfig.CertificateFile.FlagName, SrvConfig.CertificateFile.FlagDescription)
	flag.BoolVar(&SrvConfig.TLSEnabled.Value, SrvConfig.TLSEnabled.FlagName, config.DefaultServerTLSEnabled, SrvConfig.TLSEnabled.FlagDescription)

	// Initialize the application
	flag.StringVar(&DBConfig.Address.Value, DBConfig.Address.FlagName, config.DefaultDatabaseAddress, DBConfig.Address.FlagDescription)
	flag.IntVar(&DBConfig.Port.Value, DBConfig.Port.FlagName, config.DefaultDatabasePort, DBConfig.Port.FlagDescription)
	flag.StringVar(&DBConfig.Username.Value, DBConfig.Username.FlagName, config.DefaultDatabaseUsername, DBConfig.Username.FlagDescription)
	flag.StringVar(&DBConfig.Password.Value, DBConfig.Password.FlagName, config.DefaultDatabasePassword, DBConfig.Password.FlagDescription)
	flag.StringVar(&DBConfig.Name.Value, DBConfig.Name.FlagName, config.DefaultDatabaseName, DBConfig.Name.FlagDescription)
	flag.StringVar(&DBConfig.SSLMode.Value, DBConfig.SSLMode.FlagName, config.DefaultDatabaseSSLMode, DBConfig.SSLMode.FlagDescription)
	flag.DurationVar(&DBConfig.MaxPingTimeout.Value, DBConfig.MaxPingTimeout.FlagName, config.DefaultDatabaseMaxPingTimeout, DBConfig.MaxPingTimeout.FlagDescription)
	flag.DurationVar(&DBConfig.MaxQueryTimeout.Value, DBConfig.MaxQueryTimeout.FlagName, config.DefaultDatabaseMaxQueryTimeout, DBConfig.MaxQueryTimeout.FlagDescription)
	flag.DurationVar(&DBConfig.ConnMaxLifetime.Value, DBConfig.ConnMaxLifetime.FlagName, config.DefaultDatabaseConnMaxLifetime, DBConfig.ConnMaxLifetime.FlagDescription)
	flag.IntVar(&DBConfig.MaxIdleConns.Value, DBConfig.MaxIdleConns.FlagName, config.DefaultDatabaseMaxIdleConns, DBConfig.MaxIdleConns.FlagDescription)
	flag.IntVar(&DBConfig.MaxOpenConns.Value, DBConfig.MaxOpenConns.FlagName, config.DefaultDatabaseMaxOpenConns, DBConfig.MaxOpenConns.FlagDescription)

	// Parse the command line arguments
	flag.Bool("help", false, "Show this help message")
	flag.Parse()

	// implement the help flag
	if flag.Lookup("help").Value.(flag.Getter).Get().(bool) {
		flag.Usage()
		os.Exit(0)
	}

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

	slog.Debug("configuration", "database", DBConfig)
	slog.Debug("configuration", "log", LogConfig)

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Add the routes
	mux.HandleFunc("GET /version", handler.GetVersion)

	// Configure the server
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", SrvConfig.Address.Value, SrvConfig.Port.Value),
		Handler: mux,
	}

	// Configure the TLS
	if SrvConfig.TLSEnabled.Value {
		slog.Info("configuring tls")
		// if _, err := os.Stat(SrvConfig.CertificateFile.Value.Name()); os.IsNotExist(err) {
		// 	slog.Error("tls.crt file not found")
		// 	os.Exit(1)
		// }

		// if _, err := os.Stat(SrvConfig.PrivateKeyFile.Value.Name()); os.IsNotExist(err) {
		// 	slog.Error("tls.key file not found")
		// 	os.Exit(1)
		// }

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
		slog.Info("starting server", "address", SrvConfig.Address.Value, "port", SrvConfig.Port.Value)

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
