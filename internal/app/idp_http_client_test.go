//go:build unit

package app

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/slashdevops/go-rest-api-service-template/internal/config"
)

func idpClientApp(t *testing.T) *App {
	t.Helper()

	httpClient := config.NewHTTPClientConfig()
	httpClient.AllowPrivateAddresses.Value = true // the test servers are on loopback

	return &App{configs: &Configs{HTTPClient: httpClient, Authn: config.NewAuthConfig()}}
}

// TestTheIdentityProvidersClientDoesNotRetry: a sign-in is a person waiting,
// and the token POST carries a code that is good for one exchange. The
// identity providers used the shared outbound client, which retries a failed
// request ten times with backoff: a provider that did not answer held the
// callback for about two and a half minutes, and the request it retried was
// that POST.
func TestTheIdentityProvidersClientDoesNotRetry(t *testing.T) {
	t.Parallel()

	a := idpClientApp(t)
	idp := a.initIDPHTTPClient()

	if idp.Timeout != config.DefaultAuthnIDPTimeout {
		t.Errorf("timeout = %s, want authn.idp.timeout's %s", idp.Timeout, config.DefaultAuthnIDPTimeout)
	}

	t.Run("a_request_the_provider_fails_is_sent_once", func(t *testing.T) {
		t.Parallel()

		var hits atomic.Int32

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			hits.Add(1)
			w.WriteHeader(http.StatusBadGateway)
		}))
		t.Cleanup(srv.Close)

		resp, err := idp.Post(srv.URL+"/token", "application/x-www-form-urlencoded", strings.NewReader("code=one-use"))
		if err != nil {
			t.Fatalf("Post: %v", err)
		}

		resp.Body.Close()

		if got := hits.Load(); got != 1 {
			t.Errorf("the token endpoint was asked %d times, want once: the code is good for one exchange", got)
		}

		// The shared client is the contrast, and the reason for a second
		// one: it presents the same request again.
		hits.Store(0)

		shared := a.initHTTPClient()
		shared.Timeout = 5 * time.Second

		if resp, err := shared.Post(srv.URL+"/token", "application/x-www-form-urlencoded", strings.NewReader("code=one-use")); err == nil {
			resp.Body.Close()
		}

		if got := hits.Load(); got < 2 {
			t.Errorf("the shared client asked %d times: if it no longer retries, the two clients can be one again", got)
		}
	})

	t.Run("a_provider_that_is_not_there_is_an_error_at_once", func(t *testing.T) {
		t.Parallel()

		// A port nothing listens on.
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}

		addr := l.Addr().String()
		l.Close()

		started := time.Now()

		resp, err := idp.Post("http://"+addr+"/token", "application/x-www-form-urlencoded", strings.NewReader("code=one-use"))
		if err == nil {
			resp.Body.Close()
			t.Fatal("a request to a closed port succeeded")
		}

		if waited := time.Since(started); waited > 2*time.Second {
			t.Errorf("a refused connection took %s to be reported: something is retrying", waited)
		}
	})

	t.Run("the_address_guard_is_the_shared_one", func(t *testing.T) {
		t.Parallel()

		guarded := idpClientApp(t)
		guarded.configs.HTTPClient.AllowPrivateAddresses.Value = false

		srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		t.Cleanup(srv.Close)

		resp, err := guarded.initIDPHTTPClient().Get(srv.URL)
		if err == nil {
			resp.Body.Close()
			t.Fatal("the identity providers' client reached a loopback address with http.client.allow.private.addresses off")
		}
	})
}
