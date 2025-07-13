// Package server provides an HTTP server implementation with graceful shutdown and TLS support.
// It listens for OS signals to gracefully shut down or reload the server.
package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/config"
)

type HTTPServerConfig struct {
	Ctx         context.Context
	HTTPHandler http.Handler
	Config      *config.HTTPServerConfig
}

type HTTPServer struct {
	ctx        context.Context
	httpServer *http.Server
	conf       *config.HTTPServerConfig

	osSigChan chan os.Signal
	stopChan  chan struct{}

	// Protect stopChan from concurrent access
	mu       sync.Mutex
	isClosed bool
}

func NewHTTPServer(conf HTTPServerConfig) *HTTPServer {
	if conf.Ctx == nil {
		conf.Ctx = context.Background()
	}

	addr := fmt.Sprintf("%s:%d", conf.Config.Address.Value, conf.Config.Port.Value)

	ref := &HTTPServer{
		ctx: conf.Ctx,
		httpServer: &http.Server{
			Addr:    addr,
			Handler: conf.HTTPHandler,
		},
		conf:      conf.Config,
		osSigChan: make(chan os.Signal, 1),
		stopChan:  make(chan struct{}),
		isClosed:  false,
	}

	// notify the server to listen for OS signals
	signal.Notify(ref.osSigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	return ref
}

func (ref *HTTPServer) Start() {
	slog.Info("starting http server", "address", ref.httpServer.Addr, "tls", ref.conf.TLSEnabled.Value)

	// Listen for OS signals
	ref.listenOsSignals()

	if ref.conf.TLSEnabled.Value {
		if err := ref.httpServer.ListenAndServeTLS(
			ref.conf.CertificateFile.Value.Name(),
			ref.conf.PrivateKeyFile.Value.Name(),
		); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server error", "error", err)

			ref.Stop()
		}
	} else {
		if err := ref.httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server error", "error", err)

			ref.Stop()
		}
	}
}

func (ref *HTTPServer) Wait() <-chan struct{} {
	return ref.stopChan
}

func (ref *HTTPServer) Stop() {
	ref.mu.Lock()
	defer ref.mu.Unlock()

	// If already closed, don't try to send on the channel
	if ref.isClosed {
		slog.Debug("stop channel already closed")
		return
	}

	// Use a non-blocking send to avoid potential deadlocks
	select {
	case ref.stopChan <- struct{}{}:
		// Successfully sent stop signal
		slog.Debug("sent stop signal")
	default:
		// Channel is not receiving (buffer full)
		slog.Debug("stop channel not receiving")
	}
}

func (ref *HTTPServer) listenOsSignals() {
	go func() {
		slog.Info("http server listening for OS signals")

		ctx, cancel := context.WithTimeout(ref.ctx, ref.conf.ShutdownTimeout.Value)
		defer cancel()

		for {
			select {
			case sig := <-ref.osSigChan:
				slog.Debug("received OS signal", "signal", sig)

				// Handle the signal to shutdown the server or reload
				switch sig {
				case os.Interrupt, syscall.SIGINT, syscall.SIGTERM:
					slog.Warn("shutting down http server...")
					if err := ref.httpServer.Shutdown(ctx); err != nil {
						slog.Error("http server shutdown with error", "error", err)
						os.Exit(1)
					}

					// Mark channel as closed to prevent sending on closed channel
					ref.mu.Lock()
					if !ref.isClosed {
						ref.isClosed = true
						close(ref.stopChan)
					}
					ref.mu.Unlock()

					return
				case syscall.SIGHUP:
					slog.Warn("reloading http server...")
					// Reload the server
					// This is where you would reload the server

					return
				default:
					slog.Warn("unknown signal", "signal", sig)
					return
				}
			case <-ref.stopChan:
				slog.Info("received programmatic shutdown signal")
				if err := ref.httpServer.Shutdown(ctx); err != nil {
					slog.Error("http server shutdown with error", "error", err)
					os.Exit(1)
				}

				// Mark channel as closed to prevent sending on closed channel
				ref.mu.Lock()
				if !ref.isClosed {
					ref.isClosed = true
					close(ref.stopChan)
				}
				ref.mu.Unlock()

				return
			}
		}
	}()
}

// setTLSConfig sets the TLS configuration for the server.
//
//lint:ignore U1000 This function is used depending on the configuration.
func (ref *HTTPServer) setTLSConfig() error {
	slog.Info("configuring tls")
	if _, err := os.Stat(ref.conf.CertificateFile.Value.Name()); os.IsNotExist(err) {
		slog.Error(".crt file not found", "file", ref.conf.CertificateFile.Value.Name(), "error", err)
		return err
	}

	if _, err := os.Stat(ref.conf.PrivateKeyFile.Value.Name()); os.IsNotExist(err) {
		slog.Error(".key file not found", "file", ref.conf.PrivateKeyFile.Value.Name(), "error", err)
		return err
	}

	tlsCfg := &tls.Config{
		MinVersion:               tls.VersionTLS13,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}
	ref.httpServer.TLSConfig = tlsCfg
	ref.httpServer.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)

	return nil
}
