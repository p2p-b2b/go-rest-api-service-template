package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/paginator"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/service"
)

// UserHandler represents the handler for the user.
type UserHandler struct {
	service service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// RegisterRoutes registers the routes for the user.
func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /users/{id}", h.GetByID)
	mux.HandleFunc("PUT /users/{id}", h.UpdateUser)
	mux.HandleFunc("DELETE /users/{id}", h.DeleteUser)
	mux.HandleFunc("POST /users", h.CreateUser)
	mux.HandleFunc("GET /users", h.ListUsers)
}

// GetByID Get a user by ID
// @Summary Get a user by ID
// @Description Get a user by ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} model.User
// @Failure 400 {object} APIError
// @Failure 500 {object} APIError
// @Router /users/{id} [get]
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	if idString == "" {
		WriteError(w, r, http.StatusBadRequest, ErrIDRequired.Error())
		return
	}

	// convert the id to uuid.UUID
	id, err := uuid.Parse(idString)
	if err != nil {
		WriteError(w, r, http.StatusBadRequest, ErrInvalidID.Error())
		return
	}

	user, err := h.service.GetUserByID(r.Context(), id)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}

	// encode and write the response
	if err := json.NewEncoder(w).Encode(user); err != nil {
		WriteError(w, r, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// CreateUser Create a new user
// @Summary Create a new user, if the id is not provided, it will be generated
// @Description Create a new user from scratch, you should provide the id, first name, last name and email.
// @Description If the id is not provided, it will be generated automatically.
// @Tags users
// @Accept json
// @Produce json
// @Param user body model.CreateUserRequest true "CreateUserRequest"
// @Success 201 {object} string
// @Failure 400 {object} APIError
// @Failure 500 {object} APIError
// @Router /users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if user.FirstName == "" {
		WriteError(w, r, http.StatusBadRequest, "First name is required")
		return
	}

	if user.LastName == "" {
		WriteError(w, r, http.StatusBadRequest, "Last name is required")
		return
	}

	if user.Email == "" {
		WriteError(w, r, http.StatusBadRequest, "Email is required")
		return
	}

	if err := h.service.CreateUser(r.Context(), &user); err != nil {
		WriteError(w, r, http.StatusInternalServerError, ErrInternalServerError.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

// UpdateUser Update a user
// @Summary Update a user
// @Description Update a user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body model.UpdateUserRequest true "User"
// @Success 200
// @Failure 400 {object} APIError
// @Failure 500 {object} APIError
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	idParam := r.PathValue("id")
	if idParam == "" {
		WriteError(w, r, http.StatusBadRequest, ErrIDRequired.Error())
		return
	}

	// convert the id to uuid.UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// set the user ID
	user.ID = id

	// at least one field must be updated
	if user.FirstName == "" && user.LastName == "" && user.Email == "" {
		WriteError(w, r, http.StatusBadRequest, "At least one field must be updated")
		return
	}

	if err := h.service.UpdateUser(r.Context(), &user); err != nil {
		WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// DeleteUser Delete a user
// @Summary Delete a user
// @Description Delete a user
// @Tags users
// @Param id path string true "User ID"
// @Success 200
// @Failure 400 {object} APIError
// @Failure 500 {object} APIError
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idParam := r.PathValue("id")
	if idParam == "" {
		WriteError(w, r, http.StatusBadRequest, ErrIDRequired.Error())
		return
	}

	// convert the id to uuid.UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.DeleteUser(r.Context(), id); err != nil {
		WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
}

// ListUsers Return a paginated list of users
// @Summary List all users
// @Description List all users
// @Tags users
// @Produce json
// @Param sort query string false "Sort field"
// @Param filter query string false "Filter field"
// @Param fields query string false "Fields to return"
// @Param query query string false "Query string"
// @Param next_token query string false "Next cursor"
// @Param prev_token query string false "Previous cursor"
// @Param limit query int false "Limit"
// @Success 200 {object} model.ListUserResponse
// @Failure 400 {object} APIError
// @Failure 500 {object} APIError
// @Router /users [get]
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	var req model.ListUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if err != io.EOF {
			WriteError(w, r, http.StatusBadRequest, err.Error())
			return
		}
	}
	defer r.Body.Close()

	// paginator
	nextToken := r.URL.Query().Get("next_token")
	prevToken := r.URL.Query().Get("prev_token")
	limitString := r.URL.Query().Get("limit")

	sort := r.URL.Query().Get("sort")

	filterFields := r.URL.Query().Get("filter")
	var filter []string
	if len(filterFields) != 0 {
		filter = strings.Split(filterFields, ",")
	}

	fieldsFields := r.URL.Query().Get("fields")
	var fields []string
	if len(fieldsFields) != 0 {
		fields = strings.Split(fieldsFields, ",")
	}

	// convert the limit to int
	limit := paginator.DefaultLimit
	var err error
	if limitString != "" {
		limit, err = strconv.Atoi(limitString)
		if err != nil {
			WriteError(w, r, http.StatusBadRequest, "Invalid limit")
			return
		}

		if limit < 0 {
			WriteError(w, r, http.StatusBadRequest, "Limit must be greater than or equal to 0")
			return
		}
		if limit == 0 {
			limit = paginator.DefaultLimit
		}
	}

	params := &model.ListUserRequest{
		Sort:   sort,
		Filter: filter,
		Fields: fields,
		Paginator: paginator.Paginator{
			NextToken: nextToken,
			PrevToken: prevToken,
			Limit:     limit,
		},
	}

	usersResponse, err := h.service.ListUsers(r.Context(), params)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// inject the parameters and server url to paginator for the next and previous links
	srvScheme := "http"
	if r.TLS != nil {
		r.URL.Scheme = "https"
	}
	serverURL := srvScheme + "://" + r.Host

	if usersResponse.Paginator.NextToken != "" {
		usersResponse.Paginator.NextPage = serverURL + r.URL.Path + "?next_token=" + usersResponse.Paginator.NextToken + "&limit=" + strconv.Itoa(limit)
	}
	if usersResponse.Paginator.PrevToken != "" {
		usersResponse.Paginator.PrevPage = serverURL + r.URL.Path + "?prev_token=" + usersResponse.Paginator.PrevToken + "&limit=" + strconv.Itoa(limit)
	}

	// write the response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(usersResponse); err != nil {
		WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
}
