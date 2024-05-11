package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	mocksService "github.com/p2p-b2b/go-rest-api-service-template/mocks/service"
)

func TestUser_GetUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	// mockRepository := mocks.NewMockUserRepository(ctrl)
	mockService := mocksService.NewMockUserService(ctrl)
	// ctx := context.TODO()

	t.Run("GetByID", func(t *testing.T) {
		type test struct {
			name             string
			method           string
			pathPattern      string
			pathValue        string
			serviceError     error
			expectedHTTPCode int
			expectedBody     string
			mockCall         *gomock.Call
		}

		tests := []test{
			{
				name:             "page not found",
				method:           http.MethodGet,
				pathPattern:      "/users/{id}",
				pathValue:        "/users",
				serviceError:     nil,
				expectedHTTPCode: http.StatusNotFound,
				expectedBody:     "404 page not found\n",
				mockCall:         nil,
			},
			{
				name:             "invalid id",
				method:           http.MethodGet,
				pathPattern:      "/users/{id}",
				pathValue:        "/users/InvalidUUID",
				serviceError:     ErrInvalidID,
				expectedHTTPCode: http.StatusBadRequest,
				expectedBody:     ErrInvalidID.Error() + "\n",
				mockCall:         nil,
			},
			{
				name:             "service fail with internal server error",
				method:           http.MethodGet,
				pathPattern:      "/users/{id}",
				pathValue:        "/users/123e4567-e89b-12d3-a456-426614174000",
				serviceError:     ErrInternalServer,
				expectedHTTPCode: http.StatusInternalServerError,
				expectedBody:     ErrInternalServer.Error() + "\n",
				mockCall:         mockService.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).Return(nil, ErrInternalServer).Times(1),
			},
			{
				name:             "service success",
				method:           http.MethodGet,
				pathPattern:      "/users/{id}",
				pathValue:        "/users/123e4567-e89b-12d3-a456-426614174000",
				serviceError:     nil,
				expectedHTTPCode: http.StatusOK,
				expectedBody:     "{\"id\":\"ffffffff-ffff-ffff-ffff-ffffffffffff\",\"first_name\":\"\",\"last_name\":\"\",\"email\":\"\",\"created_at\":\"2021-01-01T00:00:00Z\",\"updated_at\":\"0001-01-01T00:00:00Z\"}\n",
				mockCall: mockService.
					EXPECT().
					GetUserByID(gomock.Any(), gomock.Any()).
					Return(&model.User{
						ID: uuid.Max,
						// fixed time here
						CreatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
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
				handlerPattern := fmt.Sprintf("%s %s", tc.method, tc.pathPattern)

				w := httptest.NewRecorder()

				if tc.mockCall != nil {
					gomock.InOrder(tc.mockCall)
				}

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
}
