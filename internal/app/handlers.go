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
	a.handlers.Version = handler.NewVersionHandler()

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

	// add other handlers here as needed

	return nil
}
