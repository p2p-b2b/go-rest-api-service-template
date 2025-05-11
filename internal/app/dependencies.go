package app

import (
	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/handler"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/repository"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/service"
)

// Repositories holds all repository instances
type Repositories struct {
	Health *repository.HealthRepository
	Users  *repository.UsersRepository

	// add other repositories here
}

// Services holds all service instances
type Services struct {
	Health *service.HealthService
	Users  *service.UsersService

	// add other services here
}

// Handlers holds all handler instances
type Handlers struct {
	Swagger *handler.SwaggerHandler
	Version *handler.VersionHandler
	Health  *handler.HealthHandler
	Users   *handler.UsersHandler

	// add other handlers here
}
