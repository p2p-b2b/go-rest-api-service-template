package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uuid"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/slashdevops/go-rest-api-service-template/internal/adapter/driving/http/jwtvalidator"
	"github.com/slashdevops/go-rest-api-service-template/internal/adapter/driving/http/payload"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
)

type fakeValidator struct {
	claims   map[string]any
	err      error
	gotToken string
}

func (f *fakeValidator) Validate(_ context.Context, token string) (map[string]any, error) {
	f.gotToken = token
	return f.claims, f.err
}

func (f *fakeValidator) GetClientID() string {
	return "test-client"
}

func decodeHTTPMessage(t *testing.T, rec *httptest.ResponseRecorder) payload.HTTPMessage {
	t.Helper()

	var msg payload.HTTPMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &msg); err != nil {
		t.Fatalf("expected JSON message, got error: %v", err)
	}
	return msg
}

func TestContextKeyString(t *testing.T) {
	t.Parallel()

	key := ContextKey("jwt_claims")
	if key.String() != "jwt_claims" {
		t.Fatalf("expected key string jwt_claims, got %q", key.String())
	}
}

func TestMiddlewareHelpersAndChain(t *testing.T) {
	t.Parallel()

	var calls []string

	first := Middleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "first")
			next.ServeHTTP(w, r)
		})
	})

	second := Middleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "second")
			next.ServeHTTP(w, r)
		})
	})

	h := Chain(first, second).ThenFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls = append(calls, "handler")
		w.WriteHeader(http.StatusNoContent)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	expected := []string{"first", "second", "handler"}
	if len(calls) != len(expected) {
		t.Fatalf("expected %d calls, got %d", len(expected), len(calls))
	}
	for i := range expected {
		if calls[i] != expected[i] {
			t.Fatalf("expected call %d to be %q, got %q", i, expected[i], calls[i])
		}
	}

	chained := Append(second, first)
	if len(chained) != 2 {
		t.Fatalf("expected 2 middlewares after append, got %d", len(chained))
	}

	applyHandler := first.Apply(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	recApply := httptest.NewRecorder()
	applyHandler.ServeHTTP(recApply, req)
	if recApply.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, recApply.Code)
	}

	thenHandler := first.Then(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	recThen := httptest.NewRecorder()
	thenHandler.ServeHTTP(recThen, req)
	if recThen.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, recThen.Code)
	}
}

func TestHeaderAPIVersion(t *testing.T) {
	t.Parallel()

	t.Run("default_version", func(t *testing.T) {
		t.Parallel()

		h := HeaderAPIVersion("").ThenFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		h.ServeHTTP(rec, req)

		if rec.Header().Get("X-API-Version") != "v1" {
			t.Fatalf("expected default API version v1, got %q", rec.Header().Get("X-API-Version"))
		}
	})

	t.Run("custom_version", func(t *testing.T) {
		t.Parallel()

		h := HeaderAPIVersion("v2").ThenFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		h.ServeHTTP(rec, req)

		if rec.Header().Get("X-API-Version") != "v2" {
			t.Fatalf("expected custom API version v2, got %q", rec.Header().Get("X-API-Version"))
		}
	})
}

func TestOtelTextMapPropagation(t *testing.T) {
	t.Parallel()

	old := otel.GetTextMapPropagator()
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() {
		otel.SetTextMapPropagator(old)
	})

	var traceID string
	h := OtelTextMapPropagation(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		sc := trace.SpanContextFromContext(r.Context())
		traceID = sc.TraceID().String()
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	h.ServeHTTP(rec, req)

	if traceID != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Fatalf("expected extracted trace id, got %q", traceID)
	}
}

func TestRewriteStandardErrorsAsJSON(t *testing.T) {
	t.Parallel()

	t.Run("rewrite_not_found_plain_text", func(t *testing.T) {
		t.Parallel()

		h := RewriteStandardErrorsAsJSON(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("not found"))
		}))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/missing", nil)
		h.ServeHTTP(rec, req)

		var msg payload.HTTPMessage
		if err := json.Unmarshal(rec.Body.Bytes(), &msg); err != nil {
			t.Fatalf("expected JSON body, got unmarshal error: %v", err)
		}
		if msg.StatusCode != http.StatusNotFound || msg.Message != "Not Found" {
			t.Fatalf("unexpected message content: %+v", msg)
		}
		if msg.Method != http.MethodGet || msg.Path != "/missing" {
			t.Fatalf("unexpected method/path in message: %+v", msg)
		}
	})

	t.Run("rewrite_method_not_allowed_plain_text", func(t *testing.T) {
		t.Parallel()

		h := RewriteStandardErrorsAsJSON(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte("method not allowed"))
		}))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/resource", nil)
		h.ServeHTTP(rec, req)

		var msg payload.HTTPMessage
		if err := json.Unmarshal(rec.Body.Bytes(), &msg); err != nil {
			t.Fatalf("expected JSON body, got unmarshal error: %v", err)
		}
		if msg.StatusCode != http.StatusMethodNotAllowed || msg.Message != "Method Not Allowed" {
			t.Fatalf("unexpected message content: %+v", msg)
		}
		if msg.Method != http.MethodPost || msg.Path != "/resource" {
			t.Fatalf("unexpected method/path in message: %+v", msg)
		}
	})

	t.Run("leave_existing_json_untouched", func(t *testing.T) {
		t.Parallel()

		h := RewriteStandardErrorsAsJSON(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"already-json"}`))
		}))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/existing", nil)
		h.ServeHTTP(rec, req)

		if got := rec.Body.String(); got != `{"message":"already-json"}` {
			t.Fatalf("expected original JSON body to be kept, got %q", got)
		}
	})
}

func TestCors(t *testing.T) {
	t.Parallel()

	t.Run("preflight_default_options", func(t *testing.T) {
		t.Parallel()

		called := false
		h := Cors(CorsOpts{}).ThenFunc(func(_ http.ResponseWriter, _ *http.Request) {
			called = true
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}
		if called {
			t.Fatalf("expected preflight request to return early")
		}
		if rec.Header().Get("Access-Control-Allow-Methods") == "" {
			t.Fatalf("expected default allow-methods header")
		}
		if rec.Header().Get("Access-Control-Allow-Headers") == "" {
			t.Fatalf("expected default allow-headers header")
		}
		// Wildcard origin should be set by default
		if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Fatalf("expected default wildcard origin, got %q", rec.Header().Get("Access-Control-Allow-Origin"))
		}
		// Credentials should not be set when using wildcard origin
		if rec.Header().Get("Access-Control-Allow-Credentials") != "" {
			t.Fatalf("expected no credentials header with wildcard origin, got %q", rec.Header().Get("Access-Control-Allow-Credentials"))
		}
	})

	t.Run("non_preflight_with_allowed_origin", func(t *testing.T) {
		t.Parallel()

		called := false
		h := Cors(CorsOpts{
			AllowedOrigins:   []string{"https://example.com"},
			AllowedMethods:   []string{http.MethodGet},
			AllowedHeaders:   []string{"Authorization"},
			AllowCredentials: true,
		}).ThenFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "https://example.com")
		h.ServeHTTP(rec, req)

		if !called {
			t.Fatalf("expected next handler to be called")
		}
		if rec.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
			t.Fatalf("unexpected allow-origin header: %q", rec.Header().Get("Access-Control-Allow-Origin"))
		}
		if rec.Header().Get("Access-Control-Allow-Credentials") != "true" {
			t.Fatalf("expected credentials header true")
		}
		if rec.Header().Get("Vary") != "Origin" {
			t.Fatalf("expected Vary: Origin header for non-wildcard origin, got %q", rec.Header().Get("Vary"))
		}
	})

	t.Run("wildcard_with_credentials_disabled", func(t *testing.T) {
		t.Parallel()

		// AllowCredentials=true with wildcard should be automatically disabled
		h := Cors(CorsOpts{
			AllowCredentials: true, // should be disabled because origins default to "*"
		}).ThenFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		h.ServeHTTP(rec, req)

		if rec.Header().Get("Access-Control-Allow-Credentials") != "" {
			t.Fatalf("expected credentials to be disabled with wildcard origin, got %q", rec.Header().Get("Access-Control-Allow-Credentials"))
		}
	})

	t.Run("disallowed_origin_no_header", func(t *testing.T) {
		t.Parallel()

		h := Cors(CorsOpts{
			AllowedOrigins: []string{"https://allowed.com"},
		}).ThenFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "https://evil.com")
		h.ServeHTTP(rec, req)

		if rec.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Fatalf("expected no allow-origin header for disallowed origin, got %q", rec.Header().Get("Access-Control-Allow-Origin"))
		}
	})
}

func TestWrappedResponseWriter(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	w := newWrappedResponseWriter(rec)

	if w.status != http.StatusOK {
		t.Fatalf("expected default status %d, got %d", http.StatusOK, w.status)
	}

	w.WriteHeader(http.StatusAccepted)
	if w.status != http.StatusAccepted {
		t.Fatalf("expected status %d after WriteHeader, got %d", http.StatusAccepted, w.status)
	}
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected recorder status %d, got %d", http.StatusAccepted, rec.Code)
	}

	if unwrapped := w.Unwrap(); unwrapped != rec {
		t.Fatalf("expected unwrapped writer to be original recorder")
	}
}

func TestCheckAccessToken(t *testing.T) {
	t.Parallel()

	validator := &fakeValidator{
		claims: map[string]any{"token_type": domain.TokenTypeAccess.String(), "sub": "user-1"},
	}
	validators := map[jwtvalidator.ValidatorType]jwtvalidator.Validator{
		jwtvalidator.ValidatorTypeAccessToken: validator,
	}

	t.Run("missing_authorization", func(t *testing.T) {
		t.Parallel()

		h := CheckAccessToken(validators, nil).ThenFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		h.ServeHTTP(rec, req)

		msg := decodeHTTPMessage(t, rec)
		if rec.Code != http.StatusUnauthorized || msg.Message != "Missing header: Authorization" {
			t.Fatalf("unexpected response: code=%d msg=%q", rec.Code, msg.Message)
		}
	})

	t.Run("validator_error", func(t *testing.T) {
		t.Parallel()

		// The validator's error must not reach the caller. This used to write
		// err.Error() straight into the body, which is how the jwt library's
		// own text -- "crypto/ecdsa: verification error", "could not JSON
		// decode header" -- became part of this API's published contract.
		errValidator := &fakeValidator{err: errors.New("crypto/ecdsa: verification error")}
		h := CheckAccessToken(map[jwtvalidator.ValidatorType]jwtvalidator.Validator{
			jwtvalidator.ValidatorTypeAccessToken: errValidator,
		}, nil).ThenFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer abc")
		h.ServeHTTP(rec, req)

		msg := decodeHTTPMessage(t, rec)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("unexpected code: %d", rec.Code)
		}

		if strings.Contains(msg.Message, "crypto/ecdsa") {
			t.Fatalf("the library's error text reached the caller: %q", msg.Message)
		}
	})

	t.Run("invalid_token_type", func(t *testing.T) {
		t.Parallel()

		badTypeValidator := &fakeValidator{claims: map[string]any{"token_type": domain.TokenTypeRefresh.String()}}
		h := CheckAccessToken(map[jwtvalidator.ValidatorType]jwtvalidator.Validator{
			jwtvalidator.ValidatorTypeAccessToken: badTypeValidator,
		}, nil).ThenFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer abc")
		h.ServeHTTP(rec, req)

		msg := decodeHTTPMessage(t, rec)
		if rec.Code != http.StatusUnauthorized || msg.Message != "invalid token type access or personal_access" {
			t.Fatalf("unexpected response: code=%d msg=%q", rec.Code, msg.Message)
		}
	})

	t.Run("success_claims_added_to_context", func(t *testing.T) {
		t.Parallel()

		called := false
		h := CheckAccessToken(validators, nil).ThenFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			claims, ok := r.Context().Value(JwtClaims).(map[string]any)
			if !ok {
				t.Fatalf("expected claims in context")
			}
			if claims["token_type"] != domain.TokenTypeAccess.String() {
				t.Fatalf("unexpected token type in context: %v", claims["token_type"])
			}
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer good-token")
		h.ServeHTTP(rec, req)

		if !called || rec.Code != http.StatusNoContent {
			t.Fatalf("expected middleware to call next with status 204")
		}
		if validator.gotToken != "good-token" {
			t.Fatalf("expected token to be passed to validator, got %q", validator.gotToken)
		}
	})
}

func TestCheckRefreshToken(t *testing.T) {
	t.Parallel()

	t.Run("invalid_token_type", func(t *testing.T) {
		t.Parallel()

		validator := &fakeValidator{claims: map[string]any{"token_type": domain.TokenTypeAccess.String()}}
		h := CheckRefreshToken(map[jwtvalidator.ValidatorType]jwtvalidator.Validator{
			jwtvalidator.ValidatorTypeRefreshToken: validator,
		}).ThenFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer refresh-token")
		h.ServeHTTP(rec, req)

		msg := decodeHTTPMessage(t, rec)
		if rec.Code != http.StatusUnauthorized || msg.Message != "Token type is not refresh" {
			t.Fatalf("unexpected response: code=%d msg=%q", rec.Code, msg.Message)
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		validator := &fakeValidator{claims: map[string]any{"token_type": domain.TokenTypeRefresh.String(), "sub": "u-1"}}
		called := false
		h := CheckRefreshToken(map[jwtvalidator.ValidatorType]jwtvalidator.Validator{
			jwtvalidator.ValidatorTypeRefreshToken: validator,
		}).ThenFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			claims, ok := r.Context().Value(JwtClaims).(map[string]any)
			if !ok || claims["token_type"] != domain.TokenTypeRefresh.String() {
				t.Fatalf("expected refresh token claims in context")
			}
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer refresh-token")
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent || !called {
			t.Fatalf("expected next handler to be called with status 204")
		}
	})
}

func TestCheckPasswordResetToken(t *testing.T) {
	t.Parallel()

	t.Run("invalid_token_type", func(t *testing.T) {
		t.Parallel()

		validator := &fakeValidator{claims: map[string]any{"token_type": domain.TokenTypeAccess.String()}}
		h := CheckPasswordResetToken(map[jwtvalidator.ValidatorType]jwtvalidator.Validator{
			jwtvalidator.ValidatorTypePasswordResetToken: validator,
		}).ThenFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer pwd-reset-token")
		h.ServeHTTP(rec, req)

		msg := decodeHTTPMessage(t, rec)
		if rec.Code != http.StatusUnauthorized || msg.Message != "Token type is not password_reset" {
			t.Fatalf("unexpected response: code=%d msg=%q", rec.Code, msg.Message)
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		validator := &fakeValidator{claims: map[string]any{"token_type": domain.TokenTypePasswordReset.String(), "sub": "u-2"}}
		called := false
		h := CheckPasswordResetToken(map[jwtvalidator.ValidatorType]jwtvalidator.Validator{
			jwtvalidator.ValidatorTypePasswordResetToken: validator,
		}).ThenFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			claims, ok := r.Context().Value(JwtClaims).(map[string]any)
			if !ok || claims["token_type"] != domain.TokenTypePasswordReset.String() {
				t.Fatalf("expected password reset token claims in context")
			}
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer pwd-reset-token")
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent || !called {
			t.Fatalf("expected next handler to be called with status 204")
		}
	})
}

// accountStandings answers one standing, or one error, and records what it
// was asked.
type accountStandings struct {
	standing domain.AccountStanding
	err      error
	asked    []uuid.UUID
}

func (a *accountStandings) Standing(_ context.Context, id uuid.UUID) (domain.AccountStanding, error) {
	a.asked = append(a.asked, id)

	return a.standing, a.err
}

// A token that verifies says nothing about its account. RequireActiveAccount
// is on every chain that verifies an access token but logout, and answers a
// disabled account, and one that is not there any more, 401 with the code
// token_revoked: the session is over, and that is the answer a client ends a
// session on.
func TestRequireActiveAccount(t *testing.T) {
	t.Parallel()

	account := uuid.NewV7()

	withClaims := func(claims map[string]any) *http.Request {
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		if claims == nil {
			return req
		}

		return req.WithContext(context.WithValue(req.Context(), JwtClaims, claims))
	}

	for name, c := range map[string]struct {
		claims   map[string]any
		standing domain.AccountStanding
		fault    error
		// want
		status  int
		code    string
		message string
		asked   bool
	}{
		"an_active_account_goes_on": {
			claims: map[string]any{"sub": account.String()}, standing: domain.AccountActive,
			status: http.StatusNoContent, asked: true,
		},
		"a_disabled_account_is_a_session_that_is_over": {
			claims: map[string]any{"sub": account.String()}, standing: domain.AccountDisabled,
			status: http.StatusUnauthorized, code: domain.CodeTokenRevoked, message: domain.AuthnTokenRevoked, asked: true,
		},
		"an_account_that_is_not_there_is_answered_the_same": {
			claims: map[string]any{"sub": account.String()}, standing: domain.AccountMissing,
			status: http.StatusUnauthorized, code: domain.CodeTokenRevoked, message: domain.AuthnTokenRevoked, asked: true,
		},
		// A lookup that fails is not permission to continue.
		"a_fault_fails_closed": {
			claims: map[string]any{"sub": account.String()}, fault: errors.New("the database is away"),
			status: http.StatusInternalServerError, message: "internal server error", asked: true,
		},
		// No answer is not a yes: a standing nobody gave, beside no error.
		"a_standing_nobody_gave_is_refused": {
			claims: map[string]any{"sub": account.String()}, standing: domain.AccountUnknown,
			status: http.StatusUnauthorized, code: domain.CodeTokenRevoked, message: domain.AuthnTokenRevoked, asked: true,
		},
		"no_claims":                {status: http.StatusUnauthorized, message: "Invalid or expired token"},
		"no_sub_claim":             {claims: map[string]any{"jti": "x"}, status: http.StatusUnauthorized, message: "Invalid or expired token"},
		"a_sub_that_is_no_string":  {claims: map[string]any{"sub": 12}, status: http.StatusUnauthorized, message: "Invalid or expired token"},
		"a_sub_that_is_no_account": {claims: map[string]any{"sub": "nobody"}, status: http.StatusUnauthorized, message: "Invalid or expired token"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			accounts := &accountStandings{standing: c.standing, err: c.fault}

			reached := false
			h := RequireActiveAccount(accounts).ThenFunc(func(w http.ResponseWriter, _ *http.Request) {
				reached = true

				w.WriteHeader(http.StatusNoContent)
			})

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, withClaims(c.claims))

			if rec.Code != c.status {
				t.Fatalf("status = %d, want %d. Body: %s", rec.Code, c.status, rec.Body.String())
			}

			if reached != (c.status == http.StatusNoContent) {
				t.Errorf("the handler was reached = %v with status %d", reached, rec.Code)
			}

			if c.asked != (len(accounts.asked) == 1) {
				t.Errorf("the account's standing was asked %d times, want asked = %v", len(accounts.asked), c.asked)
			}

			if c.asked && accounts.asked[0] != account {
				t.Errorf("asked about %s, want the token's account %s", accounts.asked[0], account)
			}

			if c.status == http.StatusNoContent {
				return
			}

			var body struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			}

			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("not a message: %s", rec.Body.String())
			}

			if body.Message != c.message || body.Code != c.code {
				t.Errorf("answered %q (code %q), want %q (code %q)", body.Message, body.Code, c.message, c.code)
			}

			// Which of the two it was is not the caller's to learn.
			if strings.Contains(rec.Body.String(), "disabled") || strings.Contains(rec.Body.String(), account.String()) {
				t.Errorf("the answer says more than that the session is over: %s", rec.Body.String())
			}
		})
	}
}

func TestRecovery(t *testing.T) {
	t.Parallel()

	t.Run("recovers_from_panic", func(t *testing.T) {
		t.Parallel()

		h := Recovery(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			panic("something went wrong")
		}))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		h.ServeHTTP(rec, req)

		msg := decodeHTTPMessage(t, rec)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rec.Code)
		}
		if msg.Message != "internal server error" {
			t.Fatalf("unexpected message: %q", msg.Message)
		}
	})

	t.Run("no_panic_passes_through", func(t *testing.T) {
		t.Parallel()

		called := false
		h := Recovery(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ok", nil)
		h.ServeHTTP(rec, req)

		if !called {
			t.Fatalf("expected handler to be called")
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})
}

func TestLogging(t *testing.T) {
	t.Parallel()

	called := false
	h := Logging(nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	h.ServeHTTP(rec, req)

	if !called {
		t.Fatalf("expected handler to be called")
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}
}

func TestCheckToken_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("bearer_without_token", func(t *testing.T) {
		t.Parallel()

		validator := &fakeValidator{claims: map[string]any{"token_type": domain.TokenTypeAccess.String()}}
		h := CheckAccessToken(map[jwtvalidator.ValidatorType]jwtvalidator.Validator{
			jwtvalidator.ValidatorTypeAccessToken: validator,
		}, nil).ThenFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer ")
		h.ServeHTTP(rec, req)

		msg := decodeHTTPMessage(t, rec)
		if rec.Code != http.StatusUnauthorized || msg.Message != "Token is empty" {
			t.Fatalf("unexpected response: code=%d msg=%q", rec.Code, msg.Message)
		}
	})

	t.Run("no_bearer_prefix", func(t *testing.T) {
		t.Parallel()

		h := CheckAccessToken(map[jwtvalidator.ValidatorType]jwtvalidator.Validator{}, nil).ThenFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Basic abc123")
		h.ServeHTTP(rec, req)

		msg := decodeHTTPMessage(t, rec)
		if rec.Code != http.StatusUnauthorized || msg.Message != "Authorization header must use the Bearer scheme" {
			t.Fatalf("unexpected response: code=%d msg=%q", rec.Code, msg.Message)
		}
	})

	t.Run("nil_validators", func(t *testing.T) {
		t.Parallel()

		h := CheckAccessToken(nil, nil).ThenFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer some-token")
		h.ServeHTTP(rec, req)

		msg := decodeHTTPMessage(t, rec)
		if rec.Code != http.StatusUnauthorized || msg.Message != "Unauthorized" {
			t.Fatalf("unexpected response: code=%d msg=%q", rec.Code, msg.Message)
		}
	})

	t.Run("missing_validator_type", func(t *testing.T) {
		t.Parallel()

		// Pass a map without the required validator type
		h := CheckAccessToken(map[jwtvalidator.ValidatorType]jwtvalidator.Validator{}, nil).ThenFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer some-token")
		h.ServeHTTP(rec, req)

		msg := decodeHTTPMessage(t, rec)
		if rec.Code != http.StatusUnauthorized || msg.Message != "Unauthorized" {
			t.Fatalf("unexpected response: code=%d msg=%q", rec.Code, msg.Message)
		}
	})

	t.Run("empty_claims", func(t *testing.T) {
		t.Parallel()

		validator := &fakeValidator{claims: map[string]any{}}
		h := CheckAccessToken(map[jwtvalidator.ValidatorType]jwtvalidator.Validator{
			jwtvalidator.ValidatorTypeAccessToken: validator,
		}, nil).ThenFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer some-token")
		h.ServeHTTP(rec, req)

		msg := decodeHTTPMessage(t, rec)
		if rec.Code != http.StatusUnauthorized || msg.Message != "Claims is empty" {
			t.Fatalf("unexpected response: code=%d msg=%q", rec.Code, msg.Message)
		}
	})

	t.Run("missing_token_type_claim", func(t *testing.T) {
		t.Parallel()

		validator := &fakeValidator{claims: map[string]any{"sub": "user-1"}}
		h := CheckAccessToken(map[jwtvalidator.ValidatorType]jwtvalidator.Validator{
			jwtvalidator.ValidatorTypeAccessToken: validator,
		}, nil).ThenFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer some-token")
		h.ServeHTTP(rec, req)

		msg := decodeHTTPMessage(t, rec)
		if rec.Code != http.StatusUnauthorized || msg.Message != "Token type field not found in claims" {
			t.Fatalf("unexpected response: code=%d msg=%q", rec.Code, msg.Message)
		}
	})

	t.Run("personal_access_token_accepted", func(t *testing.T) {
		t.Parallel()

		validator := &fakeValidator{claims: map[string]any{"token_type": domain.TokenTypePersonalAccess.String(), "sub": "user-1"}}
		called := false
		h := CheckAccessToken(map[jwtvalidator.ValidatorType]jwtvalidator.Validator{
			jwtvalidator.ValidatorTypeAccessToken: validator,
		}, nil).ThenFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer pa-token")
		h.ServeHTTP(rec, req)

		if !called || rec.Code != http.StatusNoContent {
			t.Fatalf("expected personal_access token to be accepted")
		}
	})
}
