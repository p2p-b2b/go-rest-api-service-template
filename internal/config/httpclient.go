package config

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"
)

var (
	ErrHTTPClientInvalidIdleConns             = errors.New("invalid max idle connections. Must be between " + strconv.Itoa(ValidHTTPClientMinIdleConns) + " and " + strconv.Itoa(ValidHTTPClientMaxIdleConns))
	ErrHTTPClientInvalidConnsPerHost          = errors.New("invalid max idle connections per host. Must be between " + strconv.Itoa(ValidHTTPClientMinIdleConnsPerHost) + " and " + strconv.Itoa(ValidHTTPClientMaxIdleConnsPerHost))
	ErrHTTPClientInvalidIdleConnTimeout       = errors.New("invalid idle connection timeout. Must be between " + ValidHTTPClientMinIdleConnTimeout.String() + " and " + ValidHTTPClientMaxIdleConnTimeout.String())
	ErrHTTPClientInvalidTLSHandshakeTimeout   = errors.New("invalid TLS handshake timeout. Must be between " + ValidHTTPClientMinTLSHandshakeTimeout.String() + " and " + ValidHTTPClientMaxTLSHandshakeTimeout.String())
	ErrHTTPClientInvalidExpectContinueTimeout = errors.New("invalid expect continue timeout. Must be between " + ValidHTTPClientMinExpectContinueTimeout.String() + " and " + ValidHTTPClientMaxExpectContinueTimeout.String())
	ErrHTTPClientInvalidDisableKeepAlives     = errors.New("invalid disable keep-alives. Must be true or false")
	ErrHTTPClientInvalidTimeout               = errors.New("invalid timeout. Must be between " + ValidHTTPClientMinTimeout.String() + " and " + ValidHTTPClientMaxTimeout.String())
	ErrHTTPClientInvalidRetryStrategy         = errors.New("invalid retry strategy. Must be one of [" + ValidHTTPClientRetryStrategies + "]")
	ErrHTTPClientInvalidMaxRetries            = errors.New("invalid max retries. Must be between " + strconv.Itoa(ValidHTTPClientMinMaxRetries) + " and " + strconv.Itoa(ValidHTTPClientMaxMaxRetries))
)

const (
	ValidHTTPClientMaxIdleConns        = 200
	ValidHTTPClientMinIdleConns        = 1
	ValidHTTPClientMaxIdleConnsPerHost = 200
	ValidHTTPClientMinIdleConnsPerHost = 1
	ValidHTTPClientMaxIdleConnTimeout  = 120 * time.Second
	ValidHTTPClientMinIdleConnTimeout  = 1 * time.Second

	ValidHTTPClientMaxTLSHandshakeTimeout   = 15 * time.Second
	ValidHTTPClientMinTLSHandshakeTimeout   = 1 * time.Second
	ValidHTTPClientMaxExpectContinueTimeout = 5 * time.Second
	ValidHTTPClientMinExpectContinueTimeout = 1 * time.Second
	ValidHTTPClientMaxTimeout               = 30 * time.Second
	ValidHTTPClientMinTimeout               = 1 * time.Second

	ValidHTTPClientMaxMaxRetries   = 15
	ValidHTTPClientMinMaxRetries   = 1
	ValidHTTPClientRetryStrategies = "exponential|fixed|jitter"

	DefaultHTTPClientMaxIdleConns        = 100
	DefaultHTTPClientMaxIdleConnsPerHost = 100
	DefaultHTTPClientIdleConnTimeout     = 90 * time.Second

	DefaultHTTPClientTLSHandshakeTimeout   = 10 * time.Second
	DefaultHTTPClientExpectContinueTimeout = 1 * time.Second
	DefaultHTTPClientDisableKeepAlives     = false
	DefaultHTTPClientTimeout               = 5 * time.Second

	DefaultHTTPClientMaxRetries    = 3
	DefaultHTTPClientRetryStrategy = "jitter"
)

type HTTPClientConfig struct {
	MaxIdleConns          Field[int]
	MaxIdleConnsPerHost   Field[int]
	IdleConnTimeout       Field[time.Duration]
	TLSHandshakeTimeout   Field[time.Duration]
	ExpectContinueTimeout Field[time.Duration]
	DisableKeepAlives     Field[bool]
	Timeout               Field[time.Duration]
	MaxRetries            Field[int]
	RetryStrategy         Field[string]
}

func NewHTTPClientConfig() *HTTPClientConfig {
	return &HTTPClientConfig{
		MaxIdleConns:          NewField("http.client.max.idle.conns", "HTTP_CLIENT_MAX_IDLE_CONNS", "Maximum number of idle connections", DefaultHTTPClientMaxIdleConns),
		MaxIdleConnsPerHost:   NewField("http.client.max.idle.conns.per.host", "HTTP_CLIENT_MAX_IDLE_CONNS_PER_HOST", "Maximum number of idle connections per host", DefaultHTTPClientMaxIdleConns),
		IdleConnTimeout:       NewField("http.client.idle.conn.timeout", "HTTP_CLIENT_IDLE_CONN_TIMEOUT", "Idle connection timeout", DefaultHTTPClientIdleConnTimeout),
		TLSHandshakeTimeout:   NewField("http.client.tls.handshake.timeout", "HTTP_CLIENT_TLS_HANDSHAKE_TIMEOUT", "TLS handshake timeout", DefaultHTTPClientTLSHandshakeTimeout),
		ExpectContinueTimeout: NewField("http.client.expect.continue.timeout", "HTTP_CLIENT_EXPECT_CONTINUE_TIMEOUT", "Expect continue timeout", DefaultHTTPClientExpectContinueTimeout),
		DisableKeepAlives:     NewField("http.client.disable.keep.alives", "HTTP_CLIENT_DISABLE_KEEP_ALIVES", "Disable keep-alives", DefaultHTTPClientDisableKeepAlives),
		Timeout:               NewField("http.client.timeout", "HTTP_CLIENT_TIMEOUT", "Timeout for HTTP requests", DefaultHTTPClientTimeout),
		MaxRetries:            NewField("http.client.max.retries", "HTTP_CLIENT_MAX_RETRIES", "Maximum number of retries for HTTP requests", DefaultHTTPClientMaxRetries),
		RetryStrategy:         NewField("http.client.retry.strategy", "HTTP_CLIENT_RETRY_STRATEGY", "Retry strategy for HTTP request. Valid values are: ["+ValidHTTPClientRetryStrategies+"]", DefaultHTTPClientRetryStrategy),
	}
}

func (c *HTTPClientConfig) ParseEnvVars() {
	c.MaxIdleConns.Value = GetEnv(c.MaxIdleConns.EnVarName, c.MaxIdleConns.Value)
	c.MaxIdleConnsPerHost.Value = GetEnv(c.MaxIdleConnsPerHost.EnVarName, c.MaxIdleConnsPerHost.Value)
	c.IdleConnTimeout.Value = GetEnv(c.IdleConnTimeout.EnVarName, c.IdleConnTimeout.Value)
	c.TLSHandshakeTimeout.Value = GetEnv(c.TLSHandshakeTimeout.EnVarName, c.TLSHandshakeTimeout.Value)
	c.ExpectContinueTimeout.Value = GetEnv(c.ExpectContinueTimeout.EnVarName, c.ExpectContinueTimeout.Value)
	c.DisableKeepAlives.Value = GetEnv(c.DisableKeepAlives.EnVarName, c.DisableKeepAlives.Value)
	c.Timeout.Value = GetEnv(c.Timeout.EnVarName, c.Timeout.Value)
	c.MaxRetries.Value = GetEnv(c.MaxRetries.EnVarName, c.MaxRetries.Value)
	c.RetryStrategy.Value = GetEnv(c.RetryStrategy.EnVarName, c.RetryStrategy.Value)
}

func (c *HTTPClientConfig) Validate() error {
	if c.MaxIdleConns.Value < ValidHTTPClientMinIdleConns || c.MaxIdleConns.Value > ValidHTTPClientMaxIdleConns {
		return ErrHTTPClientInvalidIdleConns
	}

	if c.MaxIdleConnsPerHost.Value < ValidHTTPClientMinIdleConnsPerHost || c.MaxIdleConnsPerHost.Value > ValidHTTPClientMaxIdleConnsPerHost {
		return ErrHTTPClientInvalidConnsPerHost
	}

	if c.IdleConnTimeout.Value < ValidHTTPClientMinIdleConnTimeout || c.IdleConnTimeout.Value > ValidHTTPClientMaxIdleConnTimeout {
		return ErrHTTPClientInvalidIdleConnTimeout
	}

	if c.TLSHandshakeTimeout.Value < ValidHTTPClientMinTLSHandshakeTimeout || c.TLSHandshakeTimeout.Value > ValidHTTPClientMaxTLSHandshakeTimeout {
		return ErrHTTPClientInvalidTLSHandshakeTimeout
	}

	if c.ExpectContinueTimeout.Value < ValidHTTPClientMinExpectContinueTimeout || c.ExpectContinueTimeout.Value > ValidHTTPClientMaxExpectContinueTimeout {
		return ErrHTTPClientInvalidExpectContinueTimeout
	}

	if c.Timeout.Value < ValidHTTPClientMinTimeout || c.Timeout.Value > ValidHTTPClientMaxTimeout {
		return ErrHTTPClientInvalidTimeout
	}

	if !slices.Contains(strings.Split(ValidHTTPClientRetryStrategies, "|"), c.RetryStrategy.Value) {
		return ErrHTTPClientInvalidRetryStrategy
	}

	if c.MaxRetries.Value < ValidHTTPClientMinMaxRetries || c.MaxRetries.Value > ValidHTTPClientMaxMaxRetries {
		return ErrHTTPClientInvalidMaxRetries
	}

	return nil
}
