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
	"syscall"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/config"
)

type HTTPServerConfig struct {
	Ctx         context.Context
	HttpHandler http.Handler
	Config      *config.HTTPServerConfig
}

type HTTPServer struct {
	ctx        context.Context
	httpServer *http.Server
	conf       *config.HTTPServerConfig

	osSigChan chan os.Signal
	stopChan  chan struct{}
}

func NewHTTPServer(conf HTTPServerConfig) *HTTPServer {
	if conf.Ctx == nil {
		conf.Ctx = context.Background()
	}

	addr := fmt.Sprintf("%s:%d", conf.Config.Address.Value, conf.Config.Port.Value)

	server := &HTTPServer{
		ctx: conf.Ctx,
		httpServer: &http.Server{
			Addr:    addr,
			Handler: conf.HttpHandler,
		},
		conf:      conf.Config,
		osSigChan: make(chan os.Signal, 1),
		stopChan:  make(chan struct{}),
	}

	// notify the server to listen for OS signals
	signal.Notify(server.osSigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	return server
}

func (s *HTTPServer) Start() {
	slog.Info("starting http server", "address", s.httpServer.Addr, "tls", s.conf.TLSEnabled.Value)

	// Listen for OS signals
	s.listenOsSignals()

	if s.conf.TLSEnabled.Value {
		if err := s.httpServer.ListenAndServeTLS(
			s.conf.CertificateFile.Value.Name(),
			s.conf.PrivateKeyFile.Value.Name(),
		); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server error", "error", err)

			s.Stop()
		}
	} else {
		if err := s.httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server error", "error", err)

			s.Stop()
		}
	}
}

func (s *HTTPServer) Wait() <-chan struct{} {
	return s.stopChan
}

func (s *HTTPServer) Stop() {
	s.stopChan <- struct{}{}
}

func (s *HTTPServer) listenOsSignals() {
	go func() {
		slog.Info("http server listening for OS signals")

		ctx, cancel := context.WithTimeout(s.ctx, s.conf.ShutdownTimeout.Value)
		defer cancel()

		for {
			select {
			case sig := <-s.osSigChan:
				slog.Debug("http server received OS signal", "signal", sig)

				// Handle the signal to shutdown the server or reload
				switch sig {
				case os.Interrupt, syscall.SIGINT, syscall.SIGTERM:
					slog.Warn("shutting down http server...")
					if err := s.httpServer.Shutdown(ctx); err != nil {
						slog.Error("http server shutdown with error", "error", err)
						os.Exit(1)
					}
					close(s.stopChan)
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
			case <-s.stopChan:
				return
			}
		}
	}()
}

// setTLSConfig sets the TLS configuration for the server.
//
//lint:ignore U1000 This function is used depending on the configuration.
func (s *HTTPServer) setTLSConfig() error {
	slog.Info("configuring tls")
	if _, err := os.Stat(s.conf.CertificateFile.Value.Name()); os.IsNotExist(err) {
		slog.Error(".crt file not found", "file", s.conf.CertificateFile.Value.Name(), "error", err)
		return err
	}

	if _, err := os.Stat(s.conf.PrivateKeyFile.Value.Name()); os.IsNotExist(err) {
		slog.Error(".key file not found", "file", s.conf.PrivateKeyFile.Value.Name(), "error", err)
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
	s.httpServer.TLSConfig = tlsCfg
	s.httpServer.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)

	return nil
}
