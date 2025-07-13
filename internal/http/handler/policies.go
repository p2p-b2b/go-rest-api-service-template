package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/middleware"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/respond"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

//go:generate go tool mockgen -package=mocks -destination=../../../mocks/handler/policies.go -source=policies.go PoliciesService

// PoliciesService represents the service for the policies.
type PoliciesService interface {
	List(ctx context.Context, input *model.ListPoliciesInput) (*model.ListPoliciesOutput, error)
	ListByRoleID(ctx context.Context, roleID uuid.UUID, input *model.ListPoliciesInput) (*model.ListPoliciesOutput, error)

	Create(ctx context.Context, input *model.CreatePolicyInput) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Policy, error)
	UpdateByID(ctx context.Context, input *model.UpdatePolicyInput) error
	DeleteByID(ctx context.Context, input *model.DeletePolicyInput) error

	// link/unlink policies to/from roles
	LinkRoles(ctx context.Context, input *model.LinkRolesToPolicyInput) error
	UnlinkRoles(ctx context.Context, input *model.UnlinkRolesFromPolicyInput) error
}

// PoliciesHandlerConf represents the configuration for the PoliciesHandler.
type PoliciesHandlerConf struct {
	Service       PoliciesService
	OT            *o11y.OpenTelemetry
	MetricsPrefix string
}

type policiesHandlerMetrics struct {
	handlerCalls metric.Int64Counter
}

// PoliciesHandler represents the handler for the policies.
type PoliciesHandler struct {
	service       PoliciesService
	ot            *o11y.OpenTelemetry
	metricsPrefix string
	metrics       policiesHandlerMetrics
}

// NewPoliciesHandler creates a new PoliciesHandler.
func NewPoliciesHandler(conf PoliciesHandlerConf) (*PoliciesHandler, error) {
	if conf.Service == nil {
		return nil, &model.InvalidServiceError{Message: "PoliciesService is required"}
	}

	if conf.OT == nil {
		return nil, &model.InvalidOTConfigurationError{Message: "OpenTelemetry is required"}
	}

	handler := &PoliciesHandler{
		service: conf.Service,
		ot:      conf.OT,
	}

	if conf.MetricsPrefix != "" {
		handler.metricsPrefix = strings.ReplaceAll(conf.MetricsPrefix, "-", "_")
		handler.metricsPrefix += "_"
	}

	handlerCalls, err := handler.ot.Metrics.Meter.Int64Counter(
		fmt.Sprintf("%s%s", handler.metricsPrefix, "handlers_calls_total"),
		metric.WithDescription("The number of calls to the policies handler"),
	)
	if err != nil {
		return nil, err
	}

	handler.metrics.handlerCalls = handlerCalls

	return handler, nil
}

// RegisterRoutes registers the routes on the mux.
func (ref *PoliciesHandler) RegisterRoutes(mux *http.ServeMux, middlewares ...middleware.Middleware) {
	mdw := middleware.Chain(middlewares...)

	mux.Handle("GET /policies", mdw.ThenFunc(ref.list))
	mux.Handle("GET /policies/{policy_id}", mdw.ThenFunc(ref.getByID))
	mux.Handle("POST /policies", mdw.ThenFunc(ref.create))
	mux.Handle("PUT /policies/{policy_id}", mdw.ThenFunc(ref.updateByID))
	mux.Handle("DELETE /policies/{policy_id}", mdw.ThenFunc(ref.deleteByID))

	// link/unlink roles to/from a policy
	mux.Handle("POST /policies/{policy_id}/roles", mdw.ThenFunc(ref.linkRoles))
	mux.Handle("DELETE /policies/{policy_id}/roles", mdw.ThenFunc(ref.unlinkRoles))

	// list policies by role id
	mux.Handle("GET /roles/{role_id}/policies", mdw.ThenFunc(ref.listByRoleID))
}

// getByID Get a policy by ID
//
//	@ID				019791cc-06c7-7e5b-b363-1ef381f1e832
//	@Summary		Get policy
//	@Description	Retrieve a specific policy by its unique identifier
//	@Tags			Policies
//	@Param			policy_id	path	string	true	"The policy id in UUID format"	Format(uuid)
//	@Produce		json
//	@Success		200	{object}	model.Policy
//	@Failure		400	{object}	model.HTTPMessage
//	@Failure		404	{object}	model.HTTPMessage
//	@Failure		500	{object}	model.HTTPMessage
//	@Router			/policies/{policy_id} [get]
//	@Security		AccessToken
func (ref *PoliciesHandler) getByID(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Policies.getByID")
	defer span.End()

	policyID, err := parseUUIDQueryParams(r.PathValue("policy_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Policies.getByID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	out, err := ref.service.GetByID(ctx, policyID)
	if err != nil {
		var policyNotFoundError *model.PolicyNotFoundError
		if errors.As(err, &policyNotFoundError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusNotFound, "handler.Policies.getByID")
			respond.WriteJSONMessage(w, r, http.StatusNotFound, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Policies.getByID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	if err := respond.WriteJSONData(w, http.StatusOK, out); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Policies.getByID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Policies.getByID: called", "policy.id", out.ID)
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "get policy",
		attribute.String("policy.id", out.ID.String()))
}

// create Create a new policy
//
//	@ID				019791cc-06c7-7e63-8ec9-de5b38235dbf
//	@Summary		Create policy
//	@Description	Create a new policy with specified permissions
//	@Tags			Policies
//	@Accept			json
//	@Produce		json
//	@Param			body	body		model.CreatePolicyRequest	true	"Create policy Request"
//	@Success		201		{object}	model.HTTPMessage
//	@Failure		400		{object}	model.HTTPMessage
//	@Failure		409		{object}	model.HTTPMessage
//	@Failure		500		{object}	model.HTTPMessage
//	@Router			/policies [post]
//	@Security		AccessToken
func (ref *PoliciesHandler) create(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Policies.create")
	defer span.End()

	var req model.CreatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Policies.create")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if req.ID == uuid.Nil {
		var err error
		req.ID, err = uuid.NewV7()
		if err != nil {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Policies.create")
			respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
			return
		}
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Policies.create")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.CreatePolicyInput{
		ID:              req.ID,
		Name:            req.Name,
		Description:     req.Description,
		AllowedAction:   req.AllowedAction,
		AllowedResource: req.AllowedResource,
	}

	if err := ref.service.Create(ctx, input); err != nil {
		var errNameExists *model.PolicyNameAlreadyExistsError
		var errIDExists *model.PolicyIDAlreadyExistsError
		if errors.As(err, &errNameExists) || errors.As(err, &errIDExists) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusConflict, "handler.Policies.create")
			respond.WriteJSONMessage(w, r, http.StatusConflict, e.Error())
			return
		}

		var errIDNotFound *model.ResourceIDNotFoundError
		if errors.As(err, &errIDNotFound) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusNotFound, "handler.Policies.create")
			respond.WriteJSONMessage(w, r, http.StatusNotFound, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Policies.create")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Policies.create: called", "policy.id", input.ID.String())
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusCreated, "create policy",
		attribute.String("policy.id", input.ID.String()))

	// Location header is not needed for this endpoint
	w.Header().Set("Location", fmt.Sprintf("%s/%s", r.URL.Path, input.ID.String()))
	respond.WriteJSONMessage(w, r, http.StatusCreated, model.PoliciesPolicyCreatedSuccessfully)
}

// updateByID Update a policy by ID
//
//	@ID				019791cc-06c7-7e67-9e1b-49a34edfe07c
//	@Summary		Update policy
//	@Description	Modify an existing policy by its ID
//	@Tags			Policies
//	@Accept			json
//	@Produce		json
//	@Param			policy_id	path		string						true	"The policy id in UUID format"	Format(uuid)
//	@Param			body		body		model.UpdatePolicyRequest	true	"Update policy Request"
//	@Success		200			{object}	model.HTTPMessage
//	@Failure		400			{object}	model.HTTPMessage
//	@Failure		404			{object}	model.HTTPMessage
//	@Failure		500			{object}	model.HTTPMessage
//	@Router			/policies/{policy_id} [put]
//	@Security		AccessToken
func (ref *PoliciesHandler) updateByID(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Policies.updateByID")
	defer span.End()

	policyID, err := parseUUIDQueryParams(r.PathValue("policy_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Policies.updateByID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	var req model.UpdatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Policies.updateByID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Policies.updateByID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.UpdatePolicyInput{
		ID:              policyID,
		Name:            req.Name,
		Description:     req.Description,
		AllowedAction:   req.AllowedAction,
		AllowedResource: req.AllowedResource,
	}

	if err := ref.service.UpdateByID(ctx, input); err != nil {
		var errPolicyNameExists *model.PolicyNameAlreadyExistsError
		var errPolicyIDExists *model.PolicyIDAlreadyExistsError
		if errors.As(err, &errPolicyNameExists) || errors.As(err, &errPolicyIDExists) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusConflict, "handler.Policies.updateByID")
			respond.WriteJSONMessage(w, r, http.StatusConflict, e.Error())
			return
		}

		var errPolicyNotFound *model.PolicyNotFoundError
		if errors.As(err, &errPolicyNotFound) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusNotFound, "handler.Policies.updateByID")
			respond.WriteJSONMessage(w, r, http.StatusNotFound, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Policies.updateByID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Policies.updateByID: called", "policy.id", input.ID.String())
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "update policy",
		attribute.String("policy.id", input.ID.String()))

	// Location header is not needed for this endpoint
	w.Header().Set("Location", fmt.Sprintf("%s/%s", r.URL.Path, input.ID.String()))
	respond.WriteJSONMessage(w, r, http.StatusOK, model.PoliciesPolicyUpdatedSuccessfully)
}

// deleteByID Delete a policy by ID
//
//	@ID				019791cc-06c7-7e6b-b308-a2b2cbc2aaa1
//	@Summary		Delete policy
//	@Description	Remove a policy permanently from the system
//	@Tags			Policies
//	@Param			policy_id	path	string	true	"The policy id in UUID format"	Format(uuid)
//	@Produce		json
//	@Success		200	{object}	model.HTTPMessage
//	@Failure		400	{object}	model.HTTPMessage
//	@Failure		404	{object}	model.HTTPMessage
//	@Failure		500	{object}	model.HTTPMessage
//	@Router			/policies/{policy_id} [delete]
//	@Security		AccessToken
func (ref *PoliciesHandler) deleteByID(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Policies.deleteByID")
	defer span.End()

	policyID, err := parseUUIDQueryParams(r.PathValue("policy_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Policies.deleteByID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.DeletePolicyInput{
		ID: policyID,
	}

	if err := ref.service.DeleteByID(ctx, input); err != nil {
		var policyNotFoundError *model.PolicyNotFoundError
		if errors.As(err, &policyNotFoundError) {
			// gracefully handle the case where the policy is not found
			respond.WriteJSONMessage(w, r, http.StatusOK, model.PoliciesPolicyDeletedSuccessfully)
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Policies.deleteByID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Policies.deleteByID: called", "policy.id", input.ID.String())
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "delete policy",
		attribute.String("policy.id", input.ID.String()))

	respond.WriteJSONMessage(w, r, http.StatusOK, model.PoliciesPolicyDeletedSuccessfully)
}

// list Retrieves a paginated list of all the policies in the system
//
//	@ID				019791cc-06c7-7e73-96aa-7e0383caae0d
//	@Summary		List policies
//	@Description	Retrieve paginated list of all policies in the system
//	@Tags			Policies
//	@Produce		json
//	@Param			sort		query		string	false	"Comma-separated list of fields to sort by. Example: first_name ASC, created_at DESC"	Format(string)
//	@Param			filter		query		string	false	"Filter field. Example: id=1 AND first_name='John'"										Format(string)
//	@Param			fields		query		string	false	"Fields to return. Example: id,first_name,last_name"									Format(string)
//	@Param			next_token	query		string	false	"Next cursor"																			Format(string)
//	@Param			prev_token	query		string	false	"Previous cursor"																		Format(string)
//	@Param			limit		query		int		false	"Limit"																					Format(int)
//	@Success		200			{object}	model.ListPoliciesResponse
//	@Failure		400			{object}	model.HTTPMessage
//	@Failure		500			{object}	model.HTTPMessage
//	@Router			/policies [get]
//	@Security		AccessToken
func (ref *PoliciesHandler) list(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Policies.list")
	defer span.End()

	// parse the query parameters
	params := map[string]any{
		"sort":      r.URL.Query().Get("sort"),
		"filter":    r.URL.Query().Get("filter"),
		"fields":    r.URL.Query().Get("fields"),
		"nextToken": r.URL.Query().Get("next_token"),
		"prevToken": r.URL.Query().Get("prev_token"),
		"limit":     r.URL.Query().Get("limit"),
	}

	sort, filter, fields, nextToken, prevToken, limit, err := parseListQueryParams(
		params,
		model.PoliciesPartialFields, // Corrected: Use Policy fields
		model.PoliciesFilterFields,  // Corrected: Use Policy fields
		model.PoliciesSortFields,    // Corrected: Use Policy fields
	)
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Policies.list")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.ListPoliciesInput{
		Sort:   sort,
		Filter: filter,
		Fields: fields,
		Paginator: model.Paginator{
			NextToken: nextToken,
			PrevToken: prevToken,
			Limit:     limit,
		},
	}

	out, err := ref.service.List(ctx, input)
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Policies.list")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	// Generate the next and previous pages
	location := fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
	out.Paginator.GeneratePages(location)

	if err := respond.WriteJSONData(w, http.StatusOK, out); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Policies.list")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Policies.list: called", "policies.count", len(out.Items))
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "list policy",
		attribute.Int("policies.count", len(out.Items)))
}

// linkRoles Link roles to policy
//
//	@ID				019791cc-06c7-7e77-a2c3-4ed693a2bcdd
//	@Summary		Link roles to policy
//	@Description	Associate multiple roles with a specific policy for authorization
//	@Tags			Policies,Roles
//	@Accept			json
//	@Produce		json
//	@Param			policy_id	path		string							true	"The policy id in UUID format"	Format(uuid)
//	@Param			body		body		model.LinkRolesToPolicyRequest	true	"Link policy to roles Request"
//	@Success		200			{object}	model.HTTPMessage
//	@Failure		400			{object}	model.HTTPMessage
//	@Failure		404			{object}	model.HTTPMessage
//	@Failure		500			{object}	model.HTTPMessage
//	@Router			/policies/{policy_id}/roles [post]
//	@Security		AccessToken
func (ref *PoliciesHandler) linkRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Policies.linkRoles")
	defer span.End()

	policyID, err := parseUUIDQueryParams(r.PathValue("policy_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Policies.linkRoles")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	var req model.LinkRolesToPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Policies.linkRoles")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Policies.linkRoles")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.LinkRolesToPolicyInput{
		PolicyID: policyID,
		RoleIDs:  req.RoleIDs,
	}

	if err := ref.service.LinkRoles(ctx, input); err != nil {
		var policyNotFoundError *model.PolicyNotFoundError
		if errors.As(err, &policyNotFoundError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusNotFound, "handler.Policies.linkRoles")
			respond.WriteJSONMessage(w, r, http.StatusNotFound, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Policies.linkRoles")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Policies.linkRoles: called", "policy_id", policyID.String())
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "link roles to policy",
		attribute.String("policy.id", policyID.String()))

	// Location header is not needed for this endpoint
	w.Header().Set("Location", fmt.Sprintf("%s/%s", r.URL.Path, policyID.String()))
	respond.WriteJSONMessage(w, r, http.StatusOK, model.PoliciesRolesLinkedSuccessfully)
}

// unlinkRoles Unlink roles from policy
//
//	@ID				019791cc-06c7-7e7b-bdfc-381b015c44e7
//	@Summary		Unlink roles from policy
//	@Description	Remove role associations from a specific policy
//	@Tags			Policies,Roles
//	@Accept			json
//	@Produce		json
//	@Param			policy_id	path		string								true	"The policy id in UUID format"	Format(uuid)
//	@Param			body		body		model.UnlinkRolesFromPolicyRequest	true	"Unlink policy from roles Request"
//	@Success		200			{object}	model.HTTPMessage
//	@Failure		400			{object}	model.HTTPMessage
//	@Failure		404			{object}	model.HTTPMessage
//	@Failure		500			{object}	model.HTTPMessage
//	@Router			/policies/{policy_id}/roles [delete]
//	@Security		AccessToken
func (ref *PoliciesHandler) unlinkRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Policies.unlinkRoles")
	defer span.End()

	policyID, err := parseUUIDQueryParams(r.PathValue("policy_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Policies.unlinkRoles")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	var req model.UnlinkRolesFromPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Policies.unlinkRoles")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	if err := req.Validate(); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Policies.unlinkRoles")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.UnlinkRolesFromPolicyInput{
		PolicyID: policyID,
		RoleIDs:  req.RoleIDs,
	}

	if err := ref.service.UnlinkRoles(ctx, input); err != nil {
		var policyNotFoundError *model.PolicyNotFoundError
		if errors.As(err, &policyNotFoundError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusNotFound, "handler.Policies.unlinkRoles")
			respond.WriteJSONMessage(w, r, http.StatusNotFound, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Policies.unlinkRoles")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Policies.unlinkRoles: called", "policy_id", policyID.String())
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "unlink roles from policy",
		attribute.String("policy.id", policyID.String()))

	// Location header is not needed for this endpoint
	w.Header().Set("Location", fmt.Sprintf("%s/%s", r.URL.Path, policyID.String()))
	respond.WriteJSONMessage(w, r, http.StatusOK, model.PoliciesRolesUnlinkedSuccessfully)
}

// listByRoleID List policies by role ID
//
//	@ID				019791cc-06c7-7e82-967d-e13c399f5018
//	@Summary		List policies by role
//	@Description	Retrieve paginated list of policies associated with a specific role
//	@Tags			Policies,Roles
//	@Produce		json
//	@Param			role_id		path		string	true	"The role id in UUID format"															Format(uuid)
//	@Param			sort		query		string	false	"Comma-separated list of fields to sort by. Example: first_name ASC, created_at DESC"	Format(string)
//	@Param			filter		query		string	false	"Filter field. Example: id=1 AND first_name='John'"										Format(string)
//	@Param			fields		query		string	false	"Fields to return. Example: id,first_name,last_name"									Format(string)
//	@Param			next_token	query		string	false	"Next cursor"																			Format(string)
//	@Param			prev_token	query		string	false	"Previous cursor"																		Format(string)
//	@Param			limit		query		int		false	"Limit"																					Format(int)
//	@Success		200			{object}	model.ListPoliciesResponse
//	@Failure		400			{object}	model.HTTPMessage
//	@Failure		500			{object}	model.HTTPMessage
//	@Router			/roles/{role_id}/policies [get]
//	@Security		AccessToken
func (ref *PoliciesHandler) listByRoleID(w http.ResponseWriter, r *http.Request) {
	ctx, span, metricCommonAttributes := setupContext(r, ref.ot.Traces.Tracer, "handler.Policies.listByRoleID")
	defer span.End()

	roleID, err := parseUUIDQueryParams(r.PathValue("role_id"))
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Policies.listByRoleID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	// parse the query parameters
	params := map[string]any{
		"sort":      r.URL.Query().Get("sort"),
		"filter":    r.URL.Query().Get("filter"),
		"fields":    r.URL.Query().Get("fields"),
		"nextToken": r.URL.Query().Get("next_token"),
		"prevToken": r.URL.Query().Get("prev_token"),
		"limit":     r.URL.Query().Get("limit"),
	}

	sort, filter, fields, nextToken, prevToken, limit, err := parseListQueryParams(
		params,
		model.PoliciesPartialFields,
		model.PoliciesFilterFields,
		model.PoliciesSortFields,
	)
	if err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusBadRequest, "handler.Policies.listByRoleID")
		respond.WriteJSONMessage(w, r, http.StatusBadRequest, e.Error())
		return
	}

	input := &model.ListPoliciesInput{
		Sort:   sort,
		Filter: filter,
		Fields: fields,
		Paginator: model.Paginator{
			NextToken: nextToken,
			PrevToken: prevToken,
			Limit:     limit,
		},
	}

	out, err := ref.service.ListByRoleID(ctx, roleID, input)
	if err != nil {
		var policyNotFoundError *model.PolicyNotFoundError
		if errors.As(err, &policyNotFoundError) {
			e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusNotFound, "handler.Policies.listByRoleID")
			respond.WriteJSONMessage(w, r, http.StatusNotFound, e.Error())
			return
		}

		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Policies.listByRoleID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	// Generate the next and previous pages
	location := fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
	out.Paginator.GeneratePages(location)

	if err := respond.WriteJSONData(w, http.StatusOK, out); err != nil {
		e := recordError(ctx, span, err, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusInternalServerError, "handler.Policies.listByRoleID")
		respond.WriteJSONMessage(w, r, http.StatusInternalServerError, e.Error())
		return
	}

	slog.Debug("handler.Policies.listByRoleID: called", "policies.count", len(out.Items))
	recordSuccess(ctx, span, ref.metrics.handlerCalls, metricCommonAttributes, http.StatusOK, "list policies by role ID",
		attribute.Int("policies.count", len(out.Items)),
		attribute.String("role.id", roleID.String()))
}
