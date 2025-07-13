package app

import (
	"fmt"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/handler"
)

// initHandlers initializes all HTTP handlers with their service dependencies
func (a *App) initHandlers() error {
	a.handlers = &Handlers{}
	var err error

	// Create swagger handler
	a.handlers.Swagger = handler.NewSwaggerHandler("swagger/doc.json") // Adjust path as needed

	// Create version handler (no service dependency)
	a.handlers.Version, err = handler.NewVersionHandler(handler.VersionHandlerConf{
		OT: a.telemetry,
	})
	if err != nil {
		return fmt.Errorf("could not create version handler: %w", err)
	}

	// Create health handler
	a.handlers.Health, err = handler.NewHealthHandler(handler.HealthHandlerConf{
		Service: a.services.Health,
		OT:      a.telemetry,
	})
	if err != nil {
		return fmt.Errorf("could not create health handler: %w", err)
	}

	// Create users handler
	a.handlers.Users, err = handler.NewUsersHandler(handler.UsersHandlerConf{
		Service: a.services.Users,
		OT:      a.telemetry,
	})
	if err != nil {
		return fmt.Errorf("could not create users handler: %w", err)
	}

	// Create policies handler
	a.handlers.Policies, err = handler.NewPoliciesHandler(handler.PoliciesHandlerConf{
		Service: a.services.Policies,
		OT:      a.telemetry,
	})
	if err != nil {
		return fmt.Errorf("could not create policies handler: %w", err)
	}

	// Create resources handler
	a.handlers.Resources, err = handler.NewResourcesHandler(handler.ResourcesHandlerConf{
		Service: a.services.Resources,
		OT:      a.telemetry,
	})
	if err != nil {
		return fmt.Errorf("could not create resources handler: %w", err)
	}

	// Create roles handler
	a.handlers.Roles, err = handler.NewRolesHandler(handler.RolesHandlerConf{
		Service: a.services.Roles,
		OT:      a.telemetry,
	})
	if err != nil {
		return fmt.Errorf("could not create roles handler: %w", err)
	}

	// Create authn handler
	a.handlers.Authn, err = handler.NewAuthnHandler(handler.AuthnHandlerConf{
		Service: a.services.Authn,
		OT:      a.telemetry,
	})
	if err != nil {
		return fmt.Errorf("could not create authn handler: %w", err)
	}

	// Create projects handler
	a.handlers.Projects, err = handler.NewProjectsHandler(handler.ProjectsHandlerConf{
		Service: a.services.Projects,
		OT:      a.telemetry,
	})
	if err != nil {
		return fmt.Errorf("could not create projects handler: %w", err)
	}

	// Create products handler
	a.handlers.Products, err = handler.NewProductsHandler(handler.ProductsHandlerConf{
		Service: a.services.Products,
		OT:      a.telemetry,
	})
	if err != nil {
		return fmt.Errorf("could not create products handler: %w", err)
	}

	return nil
}
