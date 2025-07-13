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

	a.repositories.Policies, err = repository.NewPoliciesRepository(
		repository.PoliciesRepositoryConfig{
			DB:              a.dbPool,
			MaxPingTimeout:  a.configs.Database.MaxPingTimeout.Value,
			MaxQueryTimeout: a.configs.Database.MaxQueryTimeout.Value,
			OT:              a.telemetry,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create policies repository: %w", err)
	}

	a.repositories.Resources, err = repository.NewResourcesRepository(
		repository.ResourcesRepositoryConfig{
			DB:              a.dbPool,
			MaxPingTimeout:  a.configs.Database.MaxPingTimeout.Value,
			MaxQueryTimeout: a.configs.Database.MaxQueryTimeout.Value,
			OT:              a.telemetry,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create resources repository: %w", err)
	}

	a.repositories.Roles, err = repository.NewRolesRepository(
		repository.RolesRepositoryConfig{
			DB:              a.dbPool,
			MaxPingTimeout:  a.configs.Database.MaxPingTimeout.Value,
			MaxQueryTimeout: a.configs.Database.MaxQueryTimeout.Value,
			OT:              a.telemetry,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create roles repository: %w", err)
	}

	a.repositories.Projects, err = repository.NewProjectsRepository(
		repository.ProjectsRepositoryConfig{
			DB:              a.dbPool,
			MaxPingTimeout:  a.configs.Database.MaxPingTimeout.Value,
			MaxQueryTimeout: a.configs.Database.MaxQueryTimeout.Value,
			OT:              a.telemetry,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create projects repository: %w", err)
	}

	a.repositories.Products, err = repository.NewProductsRepository(
		repository.ProductsRepositoryConfig{
			DB:              a.dbPool,
			MaxPingTimeout:  a.configs.Database.MaxPingTimeout.Value,
			MaxQueryTimeout: a.configs.Database.MaxQueryTimeout.Value,
			OT:              a.telemetry,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create products repository: %w", err)
	}

	return nil
}
