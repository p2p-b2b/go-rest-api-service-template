package app

import (
	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/handler"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/repository"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/service"
)

// Repositories holds all repository instances
type Repositories struct {
	Health    *repository.HealthRepository
	Users     *repository.UsersRepository
	Policies  *repository.PoliciesRepository
	Resources *repository.ResourcesRepository
	Roles     *repository.RolesRepository
	Projects  *repository.ProjectsRepository
	Products  *repository.ProductsRepository
}

// Services holds all service instances
type Services struct {
	Health    *service.HealthService
	Users     *service.UsersService
	Policies  *service.PoliciesService
	Resources *service.ResourcesService
	Roles     *service.RolesService
	Authz     *service.AuthzService
	Authn     *service.AuthnService
	Projects  *service.ProjectsService
	Products  *service.ProductsService
}

// Handlers holds all handler instances
type Handlers struct {
	Version   *handler.VersionHandler
	Health    *handler.HealthHandler
	Users     *handler.UsersHandler
	Policies  *handler.PoliciesHandler
	Resources *handler.ResourcesHandler
	Roles     *handler.RolesHandler
	Swagger   *handler.SwaggerHandler
	Authn     *handler.AuthnHandler
	Projects  *handler.ProjectsHandler
	Products  *handler.ProductsHandler
}
