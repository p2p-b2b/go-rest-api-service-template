package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/slashdevops/go-rest-api-service-template/internal/adapter/driving/http/middleware"
	"github.com/slashdevops/go-rest-api-service-template/internal/adapter/driving/http/respond"
	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
)

func TestParseUUIDQueryParams(t *testing.T) {
	testID := uuid.NewV7()

	tests := []struct {
		input    string
		expected uuid.UUID
		err      error
	}{
		{"", uuid.Nil(), &domain.InvalidUUIDError{Message: "input is empty"}},
		{"invalid-uuid", uuid.Nil(), &domain.InvalidUUIDError{UUID: "invalid-uuid", Message: "invalid UUID length: 12"}},
		{uuid.Nil().String(), uuid.Nil(), &domain.InvalidUUIDError{UUID: uuid.Nil().String(), Message: "input is nil"}},
		{testID.String(), testID, nil},
	}

	for _, test := range tests {
		result, err := parseUUIDQueryParams(test.input)
		assert.Equal(t, test.expected, result)
		assert.Equal(t, test.err, err)
	}
}

func TestNormalizeFieldsQueryParam(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"id,name", "id,name"},
		{"id, name, email", "id,name,email"},
		{"  id ,  name ", "id,name"},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, normalizeFieldsQueryParam(test.in))
	}
}

func TestHTTPStatusForListError(t *testing.T) {
	// Sort/filter/fields allow-list validation now lives in the domain layer,
	// so a bad field comes back from the use-case as a domain validation error;
	// together with the pg errors a malformed expression triggers, these are
	// client faults that must map to 400. Everything else is a 500.
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"validation_errors", &domain.ValidationErrors{Errors: []domain.ValidationError{{Field: "sort", Message: "field not allowed"}}}, http.StatusBadRequest},
		{"validation_error", &domain.ValidationError{Field: "sort", Message: "field not allowed"}, http.StatusBadRequest},
		{"invalid_byte_sequence", &domain.InvalidByteSequenceError{}, http.StatusBadRequest},
		{"invalid_message_format", &domain.InvalidMessageFormatError{}, http.StatusBadRequest},
		{"undefined_column", &domain.UndefinedColumnError{}, http.StatusBadRequest},
		{"datatype_mismatch", &domain.DatatypeMismatchError{}, http.StatusBadRequest},
		{"wrapped_validation_errors", fmt.Errorf("list failed: %w", &domain.ValidationErrors{Errors: []domain.ValidationError{{Field: "filter"}}}), http.StatusBadRequest},
		{"unknown_error", errors.New("boom"), http.StatusInternalServerError},
		{"not_found_is_not_a_list_client_fault", &domain.RoleNotFoundError{}, http.StatusInternalServerError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, httpStatusForListError(test.err))
		})
	}
}

func TestParseNextTokenQueryParams(t *testing.T) {
	testID := uuid.NewV7()
	validToken := domain.EncodeToken(testID, 10, domain.TokenDirectionNext)

	tests := []struct {
		name      string
		nextToken string
		expected  string
		err       error
	}{
		{
			name:      "Empty token",
			nextToken: "",
			expected:  "",
			err:       nil,
		},
		{
			name:      "Invalid token",
			nextToken: "invalid",
			expected:  "",
			err:       &domain.InvalidTokenError{Message: "invalid token: illegal base64 data at input byte 4"},
		},
		{
			name:      "Valid token",
			nextToken: validToken,
			expected:  validToken,
			err:       nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := parseNextTokenQueryParams(test.nextToken)

			if test.err == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.IsType(t, test.err, err)
				// assert.Contains(t, err.Error(), test.err.Error())
			}

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestParsePrevTokenQueryParams(t *testing.T) {
	testID := uuid.NewV7()
	validPrevToken := domain.EncodeToken(testID, 10, domain.TokenDirectionPrev)
	validNextToken := domain.EncodeToken(testID, 10, domain.TokenDirectionNext)

	tests := []struct {
		name      string
		prevToken string
		expected  string
		err       error
	}{
		{
			name:      "Empty token",
			prevToken: "",
			expected:  "",
			err:       nil,
		},
		{
			name:      "Invalid token (bad format)",
			prevToken: "invalid", // Input that causes a base64 decoding error
			expected:  "",
			err:       &domain.InvalidTokenError{Message: "invalid prev_token"},
		},
		{
			name:      "Valid token (encoded as Prev)",
			prevToken: validPrevToken,
			expected:  validPrevToken,
			err:       nil,
		},
		{
			name:      "Invalid token (direction mismatch - next token for prev param)",
			prevToken: validNextToken,
			expected:  "", // On error, the function returns an empty string for the token
			err:       &domain.InvalidTokenError{Message: "invalid prev_token"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := parsePrevTokenQueryParams(test.prevToken)

			if test.err == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.IsType(t, test.err, err)
				// Check if the actual error message contains the expected message
				// This provides flexibility if the error messages don't match exactly
				// but the core issue is the same.
				assert.Contains(t, err.Error(), test.err.Error())
			}

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestParseLimitQueryParams(t *testing.T) {
	tests := []struct {
		name     string
		limit    string
		expected int
		err      error
	}{
		{
			name:     "empty string",
			limit:    "",
			expected: domain.PaginatorDefaultLimit,
			err:      nil,
		},
		{
			name:     "invalid string",
			limit:    "invalid",
			expected: 0,
			err:      &domain.InvalidLimitError{MinLimit: domain.PaginatorMinLimit, MaxLimit: domain.PaginatorMaxLimit},
		},
		{
			name:     "zero returns error",
			limit:    "0",
			expected: 0,
			err:      &domain.InvalidLimitError{MinLimit: domain.PaginatorMinLimit, MaxLimit: domain.PaginatorMaxLimit},
		},
		{
			name:     "valid limit",
			limit:    "5",
			expected: 5,
			err:      nil,
		},
		{
			name:     "negative limit",
			limit:    "-1",
			expected: 0,
			err:      &domain.InvalidLimitError{MinLimit: domain.PaginatorMinLimit, MaxLimit: domain.PaginatorMaxLimit},
		},
		{
			name:     "max limit",
			limit:    "1000",
			expected: domain.PaginatorMaxLimit,
			err:      nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := parseLimitQueryParams(test.limit)
			assert.Equal(t, test.expected, result)

			if test.err == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.IsType(t, test.err, err)
			}
		})
	}
}

func TestParseListQueryParams(t *testing.T) {
	t.Run("TestParseListQueryParams", func(t *testing.T) {
		testID := uuid.NewV7()

		params := map[string]any{
			"sort":      "name ASC",
			"filter":    "status='active'",
			"fields":    "id, name",
			"nextToken": domain.EncodeToken(testID, 10, domain.TokenDirectionNext),
			"prevToken": domain.EncodeToken(testID, 10, domain.TokenDirectionPrev),
			"limit":     "5",
		}

		sort, filter, fields, nextToken, prevToken, limit, err := parseListQueryParams(params)
		assert.NoError(t, err)
		assert.Equal(t, "name ASC", sort)
		assert.Equal(t, "status='active'", filter)
		assert.Equal(t, "id,name", fields)
		assert.Equal(t, domain.EncodeToken(testID, 10, domain.TokenDirectionNext), nextToken)
		assert.Equal(t, domain.EncodeToken(testID, 10, domain.TokenDirectionPrev), prevToken)
		assert.Equal(t, 5, limit)
	})

	t.Run("passes_sort_and_filter_through_without_validating", func(t *testing.T) {
		// Allow-list validation of sort/filter/fields moved to the domain
		// layer, so the transport no longer rejects unknown fields here; it
		// passes sort/filter through verbatim and only space-normalizes fields.
		params := map[string]any{
			"sort":      "unknown_field DESC",
			"filter":    "unknown='x'",
			"fields":    "unknown, other",
			"nextToken": "",
			"prevToken": "",
			"limit":     "5",
		}

		sort, filter, fields, _, _, _, err := parseListQueryParams(params)
		assert.NoError(t, err)
		assert.Equal(t, "unknown_field DESC", sort)
		assert.Equal(t, "unknown='x'", filter)
		assert.Equal(t, "unknown,other", fields)
	})

	t.Run("rejects_invalid_limit", func(t *testing.T) {
		params := map[string]any{
			"sort":      "",
			"filter":    "",
			"fields":    "",
			"nextToken": "",
			"prevToken": "",
			"limit":     "not-a-number",
		}

		_, _, _, _, _, _, err := parseListQueryParams(params)
		assert.Error(t, err)
		assert.IsType(t, &domain.InvalidLimitError{}, err)
	})
}

func TestGetUserIDFromContext(t *testing.T) {
	testID := uuid.NewV7()

	tests := []struct {
		name        string
		setupCtx    func() context.Context
		expected    uuid.UUID
		expectedErr error
	}{
		{
			name: "missing_jwt_claims",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expected:    uuid.Nil(),
			expectedErr: &noCallerError{reason: "there are no claims in the context"},
		},
		{
			name: "invalid_claims_format",
			setupCtx: func() context.Context {
				ctx := context.Background()
				return context.WithValue(ctx, middleware.JwtClaims, "invalid-format")
			},
			expected:    uuid.Nil(),
			expectedErr: &noCallerError{reason: "the claims are not in the expected format"},
		},
		{
			name: "missing_sub_claim",
			setupCtx: func() context.Context {
				ctx := context.Background()
				claims := map[string]any{
					"exp": 1234567890,
					"iat": 1234567890,
				}
				return context.WithValue(ctx, middleware.JwtClaims, claims)
			},
			expected:    uuid.Nil(),
			expectedErr: &noCallerError{reason: "the sub claim is missing or not a string"},
		},
		{
			name: "sub_claim_not_string",
			setupCtx: func() context.Context {
				ctx := context.Background()
				claims := map[string]any{
					"sub": 123456,
					"exp": 1234567890,
				}
				return context.WithValue(ctx, middleware.JwtClaims, claims)
			},
			expected:    uuid.Nil(),
			expectedErr: &noCallerError{reason: "the sub claim is missing or not a string"},
		},
		{
			name: "invalid_uuid_format",
			setupCtx: func() context.Context {
				ctx := context.Background()
				claims := map[string]any{
					"sub": "invalid-uuid",
					"exp": 1234567890,
				}
				return context.WithValue(ctx, middleware.JwtClaims, claims)
			},
			expected:    uuid.Nil(),
			expectedErr: &noCallerError{reason: "the sub claim is not an account id"},
		},
		{
			name: "valid_user_id",
			setupCtx: func() context.Context {
				ctx := context.Background()
				claims := map[string]any{
					"sub": testID.String(),
					"exp": 1234567890,
				}
				return context.WithValue(ctx, middleware.JwtClaims, claims)
			},
			expected:    testID,
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := test.setupCtx()
			result, err := getUserIDFromContext(ctx)

			assert.Equal(t, test.expected, result)
			if test.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.IsType(t, test.expectedErr, err)
				assert.Contains(t, err.Error(), test.expectedErr.Error())

				// A wiring fault, not the caller's: no row in the status
				// table, so the fixed 500, counted as a failure. It was an
				// InvalidJWTError, the table's 401.
				assert.Equal(t, http.StatusInternalServerError, respond.StatusFor(err))
				assert.False(t, respond.IsRefusal(err), "a missing caller is counted as a refused request")
			}
		})
	}
}

func TestGetJWTExpiration(t *testing.T) {
	tests := []struct {
		name       string
		tokenStr   string
		wantErr    bool
		errMessage string
		wantTime   int64 // Unix timestamp for easy comparison
	}{
		{
			name:     "Valid JWT with expiration",
			tokenStr: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZXhwIjoxNTg5ODM0NzU5fQ.QQ_3Yinmu5YLx5mXLkDxhIgdYkCckuiGjHYrBHDLpVQ",
			wantErr:  false,
			wantTime: 1589834759, // May 18, 2020 7:12:39 PM GMT
		},
		{
			name:       "Invalid JWT format",
			tokenStr:   "not-a-valid-jwt",
			wantErr:    true,
			errMessage: "failed to parse token",
		},
		{
			name:       "JWT without expiration claim",
			tokenStr:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIn0.Gfx6VO9tcxwk6xqx9yYzSfebfeakZp5JYIgP_edcw_A",
			wantErr:    true,
			errMessage: "does not have an expiration",
		},
		{
			name:       "JWT with invalid expiration type",
			tokenStr:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZXhwIjoibm90LWEtbnVtYmVyIn0.T12ink_7E10QxZvAQ2Y6Z_OPLMk2FeSTMhyW9MKR3Z0",
			wantErr:    true,
			errMessage: "could not get expiration time from claims",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expTime, err := getJWTExpiration(tt.tokenStr)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.True(t, expTime.IsZero(), "Expected zero time for error case")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantTime, expTime.Unix(), "Expiration time does not match expected value")
				assert.False(t, expTime.IsZero(), "Expected non-zero time for success case")
			}
		})
	}
}

// TestABodyIDThatIsNotAUUIDHasAType: an id in a request body that is not a
// uuid is refused by the standard library with a bare errors.New("invalid
// uuid"), which the decoder returns as it is. Recorded like that, the access
// log names the refusal `*errors.errorString` -- 34 such lines in one
// integration run of svc-qu3ry-core, where this was found. It is a domain
// error now, and what the caller reads has not changed.
func TestABodyIDThatIsNotAUUIDHasAType(t *testing.T) {
	var body struct {
		ID uuid.UUID `json:"id"`
	}

	req := httptest.NewRequest(http.MethodPost, "/things", strings.NewReader(`{"id":"not-a-uuid"}`))

	err := decodeJSONBody(req, &body)

	invalid, ok := errors.AsType[*domain.InvalidRequestError](err)
	if !ok {
		t.Fatalf("got %T (%v), want *domain.InvalidRequestError", err, err)
	}

	// Marked as a body that could not be decoded: the caller's to fix,
	// which is what respond.IsRefusal reads.
	if _, marked := errors.AsType[*respond.DecodeError](err); !marked || !respond.IsRefusal(err) {
		t.Errorf("a body that could not be decoded is not marked as one: %T", err)
	}

	if invalid.Message != "invalid uuid" {
		t.Errorf("message %q", invalid.Message)
	}

	rec := httptest.NewRecorder()
	respond.WriteDecodeError(rec, req, err)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status %d, want 400", rec.Code)
	}

	if want := respond.DecodeErrorMessage + ": invalid uuid"; !strings.Contains(rec.Body.String(), want) {
		t.Errorf("the caller reads %s, want it to say %q as before", rec.Body.String(), want)
	}

	// A body that decodes is untouched.
	ok2 := httptest.NewRequest(http.MethodPost, "/things", strings.NewReader(`{"id":"01990000-0000-7000-8000-000000000001"}`))
	if err := decodeJSONBody(ok2, &body); err != nil {
		t.Fatalf("a uuid was refused: %v", err)
	}
}

// TestEveryBodyThatCannotBeDecodedIsMarked: a body the decoder turns down is
// the caller's to fix, whatever the decoder calls its error -- and it calls
// most of them by encoding/json's names or io's, which say nothing of the
// kind. Unmarked, a client posting malformed JSON was counted by the layer
// metrics as a failure of the service. The mark changes nothing else: the
// text, what errors.Is finds, and what the caller is answered.
func TestEveryBodyThatCannotBeDecodedIsMarked(t *testing.T) {
	type body struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}

	for name, c := range map[string]struct {
		body string
		says string
	}{
		"not_json":                  {`{"name":`, respond.DecodeErrorMessage},
		"an_empty_body":             {``, respond.DecodeErrorMessage},
		"an_unknown_field":          {`{"nope":1}`, respond.DecodeErrorMessage + ": unknown field"},
		"a_field_of_the_wrong_type": {`{"name":1}`, respond.DecodeErrorMessage + ": field 'name' has the wrong type"},
		"an_id_that_is_no_uuid":     {`{"id":"nope"}`, respond.DecodeErrorMessage + ": invalid uuid"},
		"bytes_after_the_body":      {`{"name":"a"} {"name":"b"}`, respond.DecodeErrorMessage + ": unexpected data after the JSON body"},
	} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/things", strings.NewReader(c.body))

			var dst body

			err := decodeJSONBody(req, &dst)
			if err == nil {
				t.Fatal("the body was accepted")
			}

			if _, marked := errors.AsType[*respond.DecodeError](err); !marked {
				t.Errorf("%T is not marked as a body that could not be decoded", err)
			}

			if !respond.IsRefusal(err) {
				t.Errorf("%v is counted as a failure of the service", err)
			}

			// The mark is transparent: the unmarked error says the same.
			plain := decodeJSONBodyInto(httptest.NewRequest(http.MethodPost, "/things", strings.NewReader(c.body)), &dst)
			if plain == nil || plain.Error() != err.Error() {
				t.Errorf("marked %q, unmarked %v", err, plain)
			}

			rec := httptest.NewRecorder()
			respond.WriteDecodeError(rec, req, err)

			if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), c.says) {
				t.Errorf("answered %d %s, want 400 saying %q", rec.Code, rec.Body.String(), c.says)
			}
		})
	}

	t.Run("an_empty_body_is_still_io_eof", func(t *testing.T) {
		var dst body

		err := decodeJSONBody(httptest.NewRequest(http.MethodPost, "/things", strings.NewReader("")), &dst)
		if !errors.Is(err, io.EOF) {
			t.Errorf("%v: two handlers accept an empty body by asking errors.Is(err, io.EOF)", err)
		}
	})

	t.Run("a_body_that_decodes_is_no_error", func(t *testing.T) {
		var dst body

		if err := decodeJSONBody(httptest.NewRequest(http.MethodPost, "/things", strings.NewReader(`{"name":"a"}`)), &dst); err != nil {
			t.Errorf("a body that decodes was refused: %v", err)
		}
	})
}

// TestNoHandlerReplacesTheDecodersError: a handler records the error
// decodeJSONBody returned and hands it to respond.WriteDecodeError. Four
// auth handlers threw it away and recorded a fresh "failed to decode
// request" instead, which lost what the decoder knew: a login body cut by
// the size bound was answered 400 with that sentence, where every other
// route answers 413.
func TestNoHandlerReplacesTheDecodersError(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	var callers int

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		source, err := os.ReadFile(entry.Name())
		if err != nil {
			t.Fatalf("read: %v", err)
		}

		lines := strings.Split(string(source), "\n")
		for i, line := range lines {
			if !strings.Contains(line, "decodeJSONBody(r, ") || strings.Contains(line, "func decodeJSONBody") {
				continue
			}

			callers++

			// The three lines after the call: what is recorded.
			for _, next := range lines[i+1 : min(i+4, len(lines))] {
				if strings.Contains(next, "InvalidRequestError{") {
					t.Errorf("%s:%d records a fresh error in place of the decoder's:\n%s", entry.Name(), i+2, strings.TrimSpace(next))
				}
			}
		}
	}

	if callers < 30 {
		t.Errorf("found %d callers of decodeJSONBody: the guard is reading nothing", callers)
	}
}
