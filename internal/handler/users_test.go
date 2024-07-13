package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	mocksService "github.com/p2p-b2b/go-rest-api-service-template/mocks/handler"
)

func TestUser_GetUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	// mockRepository := mocks.NewMockUserRepository(ctrl)
	mockService := mocksService.NewMockUserService(ctrl)
	// ctx := context.TODO()

	t.Run("GetByID", func(t *testing.T) {
		type test struct {
			name        string
			method      string
			pathPattern string
			pathValue   string
			apiError    APIError
			mockCall    *gomock.Call
		}

		tests := []test{
			// {
			// 	name:        "invalid id",
			// 	method:      http.MethodGet,
			// 	pathPattern: "/users/{id}",
			// 	pathValue:   "/users/InvalidUUID",
			// 	apiError: APIError{
			// 		StatusCode: http.StatusBadRequest,
			// 		Message:    "invalid ID",
			// 	},
			// 	mockCall: nil,
			// },
			// {
			// 	name:        "service fail with internal server error",
			// 	method:      http.MethodGet,
			// 	pathPattern: "/users/{id}",
			// 	pathValue:   "/users/123e4567-e89b-12d3-a456-426614174000",
			// 	apiError: APIError{
			// 		StatusCode: http.StatusInternalServerError,
			// 		Message:    "internal server error",
			// 	},
			// 	mockCall: mockService.
			// 		EXPECT().
			// 		GetUserByID(gomock.Any(), gomock.Any()).Return(nil, ErrInternalServerError).
			// 		Times(1),
			// },
			// {
			// 	name:        "service success",
			// 	method:      http.MethodGet,
			// 	pathPattern: "/users/{id}",
			// 	pathValue:   "/users/123e4567-e89b-12d3-a456-426614174000",
			// 	apiError: APIError{
			// 		StatusCode: 0,
			// 		Message:    "",
			// 	},
			// 	mockCall: mockService.
			// 		EXPECT().
			// 		GetUserByID(gomock.Any(), gomock.Any()).
			// 		Return(&model.User{
			// 			ID: uuid.Max,
			// 			// fixed time here
			// 			CreatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			// 		}, nil).
			// 		Times(1),
			// },
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				// Given
				r, err := http.NewRequest(tc.method, tc.pathValue, nil)
				if err != nil {
					t.Fatalf("could not create request: %v", err)
				}

				// build the pattern for the handler, e.g -> GET /users/{id}
				handlerPattern := fmt.Sprintf("%s %s", tc.method, tc.pathPattern)

				w := httptest.NewRecorder()

				if tc.mockCall != nil {
					gomock.InOrder(tc.mockCall)
				}

				// Create handler config
				userHandlerConf := UserHandlerConf{
					Service: mockService,
				}
				// When
				mux := http.NewServeMux()
				h := NewUserHandler(userHandlerConf)
				mux.HandleFunc(handlerPattern, h.GetByID)
				mux.ServeHTTP(w, r)

				// Then
				if tc.apiError.StatusCode != 0 {
					if w.Code != tc.apiError.StatusCode {
						t.Errorf("expected status code %d, got %d", tc.apiError.StatusCode, w.Code)
					}

					// decode the response
					var apiError APIError
					if err := json.Unmarshal(w.Body.Bytes(), &apiError); err != nil {
						t.Fatalf("could not decode response: %v", err)
					}

					if apiError.Message != tc.apiError.Message {
						t.Errorf("expected message %q, got %q", tc.apiError.Message, apiError.Message)
					}
				}
			})
		}
	})
}
