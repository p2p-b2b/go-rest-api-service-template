package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/config"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/respond"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	mocksService "github.com/p2p-b2b/go-rest-api-service-template/mocks/handler"
	gomock "go.uber.org/mock/gomock"
)

func startsWith(value int, start int) bool {
	numStr := strconv.Itoa(value)
	startStr := strconv.Itoa(start)

	return len(numStr) > 0 && strings.HasPrefix(numStr, startStr)
}

func TestUser_GetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocksService.NewMockUsersService(ctrl)
	ctx := context.TODO()

	otConfig := config.NewOpenTelemetryConfig("test", "1.0.0")
	otConfig.TraceExporter.Value = "console"
	otConfig.MetricExporter.Value = "console"

	telemetry, err := o11y.New(ctx, otConfig)
	if err != nil {
		t.Fatalf("could not create telemetry: %v", err)
	}

	if err := telemetry.Start(); err != nil {
		t.Fatalf("could not start telemetry: %v", err)
	}

	t.Run("GetByID", func(t *testing.T) {
		type test struct {
			name         string
			method       string
			pathPattern  string
			pathValue    string
			apiError     respond.HTTPMessage
			apiResponse  model.User
			plainMessage string
			plainCode    int
			mockCall     *gomock.Call
		}

		tests := []test{
			{
				name:        "invalid uid, bad request",
				method:      http.MethodGet,
				pathPattern: "/users/{user_id}",
				pathValue:   "/users/InvalidUUID",
				apiError: respond.HTTPMessage{
					Method:     "GET",
					Path:       "/users/InvalidUUID",
					StatusCode: http.StatusBadRequest,
					Message:    "invalid UUID",
				},
				mockCall: nil,
			},
			{
				name:        "nil uid, bad request",
				method:      http.MethodGet,
				pathPattern: "/users/{user_id}",
				pathValue:   "/users/" + uuid.Nil.String(),
				apiError: respond.HTTPMessage{
					Method:     "GET",
					Path:       "/users/" + uuid.Nil.String(),
					StatusCode: http.StatusBadRequest,
					Message:    "UUID cannot be nil",
				},
				mockCall: nil,
			},
			{
				name:         "empty uid, 404 page not found",
				method:       http.MethodGet,
				pathPattern:  "/users/{user_id}",
				pathValue:    "/users/",
				plainMessage: "404 page not found\n",
				plainCode:    http.StatusNotFound,
				mockCall:     nil,
			},
			{
				name:        "empty uid, 404 page not found",
				method:      http.MethodGet,
				pathPattern: "/users/{user_id}",
				pathValue:   "/users/''",
				apiError: respond.HTTPMessage{
					Method:     "GET",
					Path:       "/users/''",
					StatusCode: http.StatusBadRequest,
					Message:    "invalid UUID",
				},
				mockCall: nil,
			},
			{
				name:        "service fail with error, return internal server error",
				method:      http.MethodGet,
				pathPattern: "/users/{user_id}",
				pathValue:   "/users/e1cdf461-87c7-465f-a374-dc6bc7e962b9",
				apiError: respond.HTTPMessage{
					Method:     "GET",
					Path:       "/users/e1cdf461-87c7-465f-a374-dc6bc7e962b9",
					StatusCode: http.StatusInternalServerError,
					Message:    "internal server error",
				},
				mockCall: mockService.
					EXPECT().
					GetByID(
						gomock.Any(),
						uuid.Must(uuid.Parse("e1cdf461-87c7-465f-a374-dc6bc7e962b9")),
					).
					Return(nil, ErrInternalServerError).
					Times(1),
			},
			{
				name:        "service success",
				method:      http.MethodGet,
				pathPattern: "/users/{user_id}",
				pathValue:   "/users/e1cdf461-87c7-465f-a374-dc6bc7e962b9",
				apiResponse: model.User{
					ID:        uuid.Must(uuid.Parse("e1cdf461-87c7-465f-a374-dc6bc7e962b9")),
					FirstName: "John",
					LastName:  "Doe",
					Email:     "jonh.doe@mail.com",
					CreatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				mockCall: mockService.
					EXPECT().
					GetByID(
						gomock.Any(),
						uuid.Must(uuid.Parse("e1cdf461-87c7-465f-a374-dc6bc7e962b9")),
					).
					Return(
						&model.User{
							ID:        uuid.Must(uuid.Parse("e1cdf461-87c7-465f-a374-dc6bc7e962b9")),
							FirstName: "John",
							LastName:  "Doe",
							Email:     "jonh.doe@mail.com",
							CreatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
							UpdatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
						}, nil).
					Times(1),
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				// Given
				r, err := http.NewRequest(tc.method, tc.pathValue, nil)
				if err != nil {
					t.Fatalf("could not create request: %v", err)
				}

				// build the pattern for the handler, e.g -> GET /users/{user_id}
				handlerPattern := fmt.Sprintf("%s %s", tc.method, tc.pathPattern)

				w := httptest.NewRecorder()

				if tc.mockCall != nil {
					gomock.InOrder(tc.mockCall)
				}

				// Create handler config
				userHandlerConf := UsersHandlerConf{
					Service: mockService,
					OT:      telemetry,
				}

				// When
				mux := http.NewServeMux()
				h, err := NewUsersHandler(userHandlerConf)
				if err != nil {
					t.Fatalf("could not create user handler: %v", err)
				}
				mux.HandleFunc(handlerPattern, h.getByID)
				mux.ServeHTTP(w, r)

				// Then
				t.Logf("status code = %d", w.Code)
				if !startsWith(w.Code, 2) {

					// when plain message is set, we don't expect a JSON response
					if tc.apiError == (respond.HTTPMessage{}) {
						if w.Code != tc.plainCode {
							t.Errorf("expected status code %d, got %d", tc.plainCode, w.Code)
						}

						if w.Body.String() != tc.plainMessage {
							t.Errorf("expected message %q, got %q", tc.plainMessage, w.Body.String())
						}

						return
					}

					if w.Code != tc.apiError.StatusCode {
						t.Errorf("expected status code %d, got %d", tc.apiError.StatusCode, w.Code)
					}

					// decode the response
					var apiError respond.HTTPMessage
					if err := json.Unmarshal(w.Body.Bytes(), &apiError); err != nil {
						t.Logf("body = %s", w.Body.Bytes())

						if w.Body.String() == tc.plainMessage {
							return
						}

						t.Fatalf("could not decode response: %v", err)
					}

					if apiError.Message != tc.apiError.Message {
						t.Log(apiError)
						t.Errorf("expected message %q, got %q", tc.apiError.Message, apiError.Message)
					}

					if apiError.Method != tc.apiError.Method {
						t.Errorf("expected method %q, got %q", tc.apiError.Method, apiError.Method)
					}

					if apiError.Path != tc.apiError.Path {
						t.Errorf("expected path %q, got %q", tc.apiError.Path, apiError.Path)
					}
				}

				if startsWith(w.Code, 2) {
					t.Logf("body = %s", w.Body.String())

					if w.Code == tc.plainCode {
						t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
					}

					var user model.User
					if err := json.Unmarshal(w.Body.Bytes(), &user); err != nil {
						t.Fatalf("could not decode response: %v", err)
					}

					if diff := cmp.Diff(tc.apiResponse, user); diff != "" {
						t.Errorf("unexpected response (-want +got):\n%s", diff)
					}

				}
			})
		}
	})
}
