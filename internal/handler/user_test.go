package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	mocksService "github.com/p2p-b2b/go-service-template/mocks/service"
)

func TestUser_GetUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	// mockRepository := mocks.NewMockUserRepository(ctrl)
	mockService := mocksService.NewMockUserService(ctrl)
	// ctx := context.TODO()

	t.Run("GetByID error codes", func(t *testing.T) {
		type test struct {
			name             string
			method           string
			pathPattern      string
			pathValue        string
			serviceError     error
			expectedHTTPCode int
			expectedBody     string
		}

		tests := []test{
			{
				name:             "id required",
				method:           http.MethodGet,
				pathPattern:      "/users/{id}",
				pathValue:        "/users",
				serviceError:     ErrInvalidID,
				expectedHTTPCode: http.StatusBadRequest,
				expectedBody:     ErrInvalidID.Error() + "\n",
			},
			{
				name:             "invalid id",
				method:           http.MethodGet,
				pathPattern:      "/users/{id}",
				pathValue:        "/users/InvalidUUID",
				serviceError:     ErrInvalidID,
				expectedHTTPCode: http.StatusBadRequest,
				expectedBody:     ErrInvalidID.Error() + "\n",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				// Given
				r, err := http.NewRequest(tc.method, tc.pathValue, nil)
				if err != nil {
					t.Fatalf("could not create request: %v", err)
				}
				handlerPattern := fmt.Sprintf("%s %s", tc.method, tc.pathPattern)

				w := httptest.NewRecorder()

				// When
				mux := http.NewServeMux()
				h := NewUserHandler(&UserHandlerConfig{
					Service: mockService,
				})

				mux.HandleFunc(handlerPattern, h.GetByID)

				mux.ServeHTTP(w, r)

				// Then
				if w.Code != tc.expectedHTTPCode {
					t.Errorf("expected status code %d, got %d", tc.expectedHTTPCode, w.Code)
				}

				if w.Body.String() != tc.expectedBody {
					t.Errorf("expected body %s, got %s", tc.expectedBody, w.Body.String())
				}
			})
		}
	})

	t.Run("get user invalid id", func(t *testing.T) {
		//  Given
		r, err := http.NewRequest(http.MethodGet, "/users/InvalidId123", nil)
		if err != nil {
			t.Fatalf("could not create request: %v", err)
		}

		w := httptest.NewRecorder()

		//  Then
		mux := http.NewServeMux()
		h := NewUserHandler(&UserHandlerConfig{
			Service: mockService,
		})

		mux.HandleFunc("/users/{id}", h.GetByID)

		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}

		if w.Body.String() != ErrInvalidID.Error()+"\n" {
			t.Errorf("expected body %s, got %s", ErrInvalidID.Error()+"\n", w.Body.String())
		}
	})
}

// 	t.Run("get user invalid id", func(t *testing.T) {
// 		// create the serve mux
// 		mux := http.NewServeMux()

// 		// create the handler
// 		h := &UserHandler{
// 			service: service,
// 		}
// 		mux.HandleFunc("GET /users/{id}", h.GetByID)

// 		// mock the request
// 		r := httptest.NewRequest(http.MethodGet, "/users/InvalidId123", nil)

// 		// mock the response writer
// 		w := httptest.NewRecorder()

// 		// serve the request
// 		mux.ServeHTTP(w, r)

// 		// assert the response
// 		if w.Code != http.StatusBadRequest {
// 			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, w.Code)
// 		}

// 		// assert the response body
// 		if w.Body.String() != ErrInvalidID.Error()+"\n" {
// 			t.Errorf("expected body %s, got %s", ErrInvalidID.Error()+"\n", w.Body.String())
// 		}
// 	})
// }
