package app

import (
	"context"
	"fmt"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/service"
)

// initServices initializes all service components of the application
func (a *App) initServices(ctx context.Context) error {
	a.services = &Services{}
	var err error

	// Health service
	a.services.Health, err = service.NewHealthService(service.HealthServiceConf{
		Repository: a.repositories.Health,
		OT:         a.telemetry,
	})
	if err != nil {
		return fmt.Errorf("error creating health service: %w", err)
	}

	// Users service
	a.services.Users, err = service.NewUsersService(service.UsersServiceConf{
		Repository: a.repositories.Users,
		OT:         a.telemetry,
	})
	if err != nil {
		return fmt.Errorf("error creating users service: %w", err)
	}

	// add other services here as needed

	return nil
}
