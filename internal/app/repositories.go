package app

import (
	"fmt"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/repository"
)

// initRepositories initializes all repositories
func (a *App) initRepositories() error {
	a.repositories = &Repositories{}
	var err error

	a.repositories.Health, err = repository.NewHealthRepository(
		repository.HealthRepositoryConfig{
			DB:             a.dbPool,
			MaxPingTimeout: a.configs.Database.MaxPingTimeout.Value,
			OT:             a.telemetry,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create health repository: %w", err)
	}

	a.repositories.Users, err = repository.NewUsersRepository(
		repository.UsersRepositoryConfig{
			DB:              a.dbPool,
			MaxPingTimeout:  a.configs.Database.MaxPingTimeout.Value,
			MaxQueryTimeout: a.configs.Database.MaxQueryTimeout.Value,
			OT:              a.telemetry,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create users repository: %w", err)
	}

	// add other repositories here as needed

	return nil
}
