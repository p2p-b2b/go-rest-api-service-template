//go:build unit

package app

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/slashdevops/go-rest-api-service-template/internal/adapter/driving/http/respond"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/o11y"
)

// TestARefusalIsWhatTheServiceAnswersBelow500: the layer metrics call an
// error a refusal when the HTTP adapter answers it with a 4xx, and a failure
// otherwise. The table of errors and statuses (respond.StatusFor) is the
// only list, with the bodies that could not be decoded beside it; this
// holds the classifier to both, on both sides of the line.
func TestARefusalIsWhatTheServiceAnswersBelow500(t *testing.T) {
	t.Parallel()

	notFound := &domain.ProjectNotFoundError{}

	for name, c := range map[string]struct {
		err     error
		refusal bool
	}{
		"a_row_that_is_not_there":        {notFound, true},
		"wrapped_by_a_layer_above":       {fmt.Errorf("getting the thing: %w", notFound), true},
		"a_bad_field":                    {&domain.ValidationErrors{}, true},
		"a_value_the_request_got_wrong":  {&domain.InvalidInputError{Message: "filter"}, true},
		"a_wrong_password":               {&domain.InvalidCredentialsError{}, true},
		"it_already_exists":              {&domain.UserAlreadyExistsError{}, true},
		"a_limit_reached":                {&domain.ResourcesLimitsHardLimitReachedError{}, true},
		"too_many_attempts":              {&domain.TooManyLoginAttemptsError{}, true},
		"a_counter_that_does_not_verify": {&domain.ResourcesLimitsInvalidSignatureError{}, false},
		"the_idp_did_not_answer_503":     {&domain.IDPUnreachableError{}, false},
		"the_driver_s_own_error":         {&pgconn.PgError{Code: "08006"}, false},
		"an_error_with_no_type":          {errors.New("connection refused"), false},
		// A body that could not be decoded is answered 400 by its own
		// helper and is in no table: the decoder's errors are not ours.
		"a_request_that_contradicts_itself": {&domain.InvalidRequestError{Message: "x"}, true},
		"a_page_token_that_is_not_one":      {&domain.InvalidTokenError{Message: "invalid next_token"}, true},
		"a_path_id_that_is_not_a_uuid":      {&domain.InvalidUUIDError{Message: "input is empty"}, true},
		// Built by a payload's Validate, answered 400 in a branch that
		// names no type.
		"an_update_that_sets_nothing":     {&domain.InvalidUserUpdateError{}, true},
		"a_body_that_is_not_json":         {&respond.DecodeError{Err: errors.New("invalid character 'x'")}, true},
		"an_empty_body":                   {&respond.DecodeError{Err: io.EOF}, true},
		"a_body_cut_by_the_size_bound":    {&respond.DecodeError{Err: &http.MaxBytesError{Limit: 1}}, true},
		"the_same_decoder_error_unmarked": {errors.New("invalid character 'x'"), false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := respond.IsRefusal(c.err); got != c.refusal {
				t.Errorf("respond.IsRefusal(%T) = %v, want %v", c.err, got, c.refusal)
			}
		})
	}
}

// TestTheClassifierIsWiredWithTelemetry: unset, every error at every layer
// is a failure and the dashboards' error ratios count refusals again. It is
// set where telemetry starts, before anything records.
func TestTheClassifierIsWiredWithTelemetry(t *testing.T) {
	// Not parallel: it sets the process-wide classifier.
	source, err := os.ReadFile("telemetry.go")
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	body := string(source)

	start := strings.Index(body, "func (a *App) initTelemetry(")
	if start < 0 {
		t.Fatal("initTelemetry is gone")
	}

	end := strings.Index(body[start:], "\n}\n")
	if !strings.Contains(body[start:start+end], "o11y.SetRefusalClassifier(respond.IsRefusal)") {
		t.Error("initTelemetry no longer hands the layer metrics their refusal classifier")
	}

	// And handed over, it is what Outcome answers with.
	o11y.SetRefusalClassifier(respond.IsRefusal)
	t.Cleanup(func() { o11y.SetRefusalClassifier(nil) })

	if got := o11y.Outcome(&domain.ProjectNotFoundError{}); got != o11y.OutcomeRefused {
		t.Errorf("Outcome(not found) = %q, want %q", got, o11y.OutcomeRefused)
	}

	if got := o11y.Outcome(errors.New("connection refused")); got != o11y.OutcomeFailed {
		t.Errorf("Outcome(untyped) = %q, want %q", got, o11y.OutcomeFailed)
	}
}
