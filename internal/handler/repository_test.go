package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	mocks "github.com/p2p-b2b/go-service-template/mocks/repository"
)

func TestRepositoryHandler_GetUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("InvalidId", func(t *testing.T) {
		// mock the repository
		repo := mocks.NewMockUserRepository(ctrl)
		mux := http.NewServeMux()

		// create the handler
		h := &RepositoryHandler{
			Repository: repo,
		}
		mux.HandleFunc("GET /users/{id}", h.GetUserByID)

		// mock the request
		r := httptest.NewRequest(http.MethodGet, "/users/InvalidId123", nil)

		// mock the response writer
		w := httptest.NewRecorder()

		// serve the request
		mux.ServeHTTP(w, r)

		// assert the response
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}

		// assert the response body
		if w.Body.String() != ErrInvalidID.Error()+"\n" {
			t.Errorf("expected body %s, got %s", ErrInvalidID.Error()+"\n", w.Body.String())
		}
	})
}
