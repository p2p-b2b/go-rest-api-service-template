//go:build integration

package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// answering is a server that answers its first `refusals` requests with the
// status, body and Retry-After it is given, and every later one with 200. It
// counts what it was sent, and keeps the last body it received.
type answering struct {
	server   *httptest.Server
	requests atomic.Int32
	lastBody atomic.Value
}

func newAnswering(t *testing.T, refusals int32, status int, body, retryAfter string) *answering {
	t.Helper()

	a := &answering{}

	a.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := a.requests.Add(1)

		sent := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(sent)
		a.lastBody.Store(string(sent))

		w.Header().Set("Content-Type", "application/json")

		if n <= refusals {
			if retryAfter != "" {
				w.Header().Set("Retry-After", retryAfter)
			}

			w.WriteHeader(status)
			_, _ = w.Write([]byte(body))

			return
		}

		_, _ = w.Write([]byte(`{"message":"ok"}`))
	}))

	t.Cleanup(a.server.Close)

	return a
}

func (a *answering) endpoint(t *testing.T, method string) *apiEndpoint {
	t.Helper()

	u, err := url.Parse(a.server.URL + "/anything")
	require.NoError(t, err)

	return &apiEndpoint{apiURL: u, requestURL: u, path: "/anything", method: method}
}

// The whole suite is one address. Its parallel phase can spend the per-IP
// budget (100 a second, burst 300), and a test that sends a handful of
// requests was answered 429 at the suite's peak: twice in a dozen gates, on
// a logout. That refusal says nothing about the route under test.
//
// sendHTTPRequest, which every test sends through, waits that refusal out
// and sends the request again. The throttle's 429 is what some tests ask
// for, and is never waited out.
//
// The limiter's own store failing is waited out too, within bounds. It was
// handed to the test at once, as a fault a test has to see; then a store
// call that ran past its 70 ms deadline on a busy machine opened the
// breaker for a few seconds, and whichever test was polling in them failed:
// three gate runs in one afternoon, none for the change under test. One
// request waits it out three times at most, and the whole run has an
// allowance: a store that is really down spends it in seconds, and from
// then on the fault reaches every test at once.
//
// Against a server of the test's own, so that what is waited out is exact.
func TestTheSuitesRequestsWaitTheBudgetOut(t *testing.T) {
	t.Parallel()

	const (
		byTheBudget = `{"message":"too many requests","code":"RATE_LIMIT_EXCEEDED","status_code":429}`
		byThrottle  = `{"message":"too many attempts, try again later","status_code":429}`
		byTheStore  = `{"message":"rate limiting is temporarily unavailable, please retry shortly","code":"RATE_LIMIT_UNAVAILABLE","status_code":429}`
	)

	t.Run("a_refusal_by_the_budget_is_waited_out_and_the_request_sent_again", func(t *testing.T) {
		t.Parallel()

		a := newAnswering(t, 2, http.StatusTooManyRequests, byTheBudget, "1")
		start := time.Now()

		response, err := sendHTTPRequest(t, t.Context(), a.endpoint(t, http.MethodPost), map[string]any{"name": "sent again as it was"})
		require.NoError(t, err)

		body := readResponseBody(t, response)
		response.Body.Close()

		assert.Equal(t, http.StatusOK, response.StatusCode, body)
		assert.EqualValues(t, 3, a.requests.Load(), "two refusals, then the answer")
		assert.GreaterOrEqual(t, time.Since(start), 2*time.Second, "the wait the answer named was not kept")
		assert.JSONEq(t, `{"name":"sent again as it was"}`, a.lastBody.Load().(string), "the request sent again is not the request")
	})

	t.Run("with_no_retry_after_it_waits_a_second", func(t *testing.T) {
		t.Parallel()

		a := newAnswering(t, 1, http.StatusTooManyRequests, byTheBudget, "")
		start := time.Now()

		response, err := sendHTTPRequest(t, t.Context(), a.endpoint(t, http.MethodGet), nil)
		require.NoError(t, err)
		response.Body.Close()

		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.GreaterOrEqual(t, time.Since(start), time.Second)
	})

	// What a test asks for, or has to see, reaches it at once.
	for name, c := range map[string]struct {
		status int
		body   string
	}{
		"the_throttles_429_reaches_the_test":      {http.StatusTooManyRequests, byThrottle},
		"a_429_that_is_not_json_reaches_the_test": {http.StatusTooManyRequests, `too many requests`},
		"the_code_under_another_status_is_no_429": {http.StatusForbidden, byTheBudget},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			a := newAnswering(t, 100, c.status, c.body, "1")
			start := time.Now()

			response, err := sendHTTPRequest(t, t.Context(), a.endpoint(t, http.MethodGet), nil)
			require.NoError(t, err)

			body := readResponseBody(t, response)
			response.Body.Close()

			assert.Equal(t, c.status, response.StatusCode)
			assert.Equal(t, c.body, body, "the answer the test reads is not the answer")
			assert.EqualValues(t, 1, a.requests.Load(), "it was sent again")
			assert.Less(t, time.Since(start), time.Second, "it was waited for")
		})
	}

	// The limiter's store could not count the request: waited out, out of an
	// allowance of the test's own here, so that the run's is not spent.
	storeFault := func(t *testing.T, refusals int32, allowance int32) (status int, requests int32, took time.Duration, left int32) {
		t.Helper()

		a := newAnswering(t, refusals, http.StatusTooManyRequests, byTheStore, "1")

		var faults atomic.Int32
		faults.Store(allowance)

		start := time.Now()

		response, err := sendHTTPRequestWaiting(t, t.Context(), &faults, a.endpoint(t, http.MethodGet), nil)
		require.NoError(t, err)
		response.Body.Close()

		return response.StatusCode, a.requests.Load(), time.Since(start), faults.Load()
	}

	t.Run("a_store_that_could_not_count_is_waited_out_and_the_request_sent_again", func(t *testing.T) {
		t.Parallel()

		status, requests, took, left := storeFault(t, 2, 10)

		assert.Equal(t, http.StatusOK, status)
		assert.EqualValues(t, 3, requests, "two refusals, then the answer")
		assert.GreaterOrEqual(t, took, 2*time.Second, "the wait the answer named was not kept")
		assert.EqualValues(t, 8, left, "each wait is taken from the run's allowance")
	})

	t.Run("one_request_waits_it_out_three_times_and_no_more", func(t *testing.T) {
		t.Parallel()

		status, requests, _, left := storeFault(t, 100, 10)

		assert.Equal(t, http.StatusTooManyRequests, status, "a store that stays down reaches the test")
		assert.EqualValues(t, storeWaits+1, requests)
		assert.EqualValues(t, 10-storeWaits, left)
	})

	t.Run("with_the_runs_allowance_spent_the_fault_reaches_the_test_at_once", func(t *testing.T) {
		t.Parallel()

		status, requests, took, _ := storeFault(t, 100, 0)

		assert.Equal(t, http.StatusTooManyRequests, status)
		assert.EqualValues(t, 1, requests, "it was sent again with nothing left to wait with")
		assert.Less(t, took, time.Second, "it was waited for")
	})

	t.Run("the_last_of_the_allowance_is_one_wait", func(t *testing.T) {
		t.Parallel()

		status, requests, _, _ := storeFault(t, 100, 1)

		assert.Equal(t, http.StatusTooManyRequests, status)
		assert.EqualValues(t, 2, requests)
	})

	// A long window is not waited for: five seconds at most per try, so a
	// budget that will not come back fails the test instead of stalling it.
	t.Run("a_long_retry_after_is_cut_to_five_seconds", func(t *testing.T) {
		t.Parallel()

		a := newAnswering(t, 1, http.StatusTooManyRequests, byTheBudget, "3600")
		start := time.Now()

		response, err := sendHTTPRequest(t, t.Context(), a.endpoint(t, http.MethodGet), nil)
		require.NoError(t, err)
		response.Body.Close()

		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Less(t, time.Since(start), 8*time.Second)
	})

	// And a request whose test is over does not keep waiting.
	t.Run("a_cancelled_context_ends_the_wait_with_the_refusal", func(t *testing.T) {
		t.Parallel()

		a := newAnswering(t, 100, http.StatusTooManyRequests, byTheBudget, "5")

		ctx, cancel := context.WithTimeout(t.Context(), 300*time.Millisecond)
		defer cancel()

		start := time.Now()

		response, err := sendHTTPRequest(t, ctx, a.endpoint(t, http.MethodGet), nil)
		require.NoError(t, err)
		response.Body.Close()

		assert.Equal(t, http.StatusTooManyRequests, response.StatusCode)
		assert.Less(t, time.Since(start), 2*time.Second)
	})
}
