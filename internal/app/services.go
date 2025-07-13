package app

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/opa"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/service"
	"github.com/p2p-b2b/mailer"
	"github.com/valkey-io/valkey-go"
)

// initServices initializes all service components of the application
func (a *App) initServices(ctx context.Context) error {
	a.services = &Services{}

	// Initialize cache client if enabled
	var cacheClient valkey.Client
	var err error

	if a.configs.Cache.Enabled.Value {
		cacheClient, err = a.initCacheClient()
		if err != nil {
			return fmt.Errorf("failed to initialize cache client: %w", err)
		}
	}

	// Create the common cache service to be used by other services
	var cacheService *service.CacheService
	if a.configs.Cache.Enabled.Value {
		cacheService = service.NewCacheService(service.CacheServiceConf{
			Cache:        cacheClient,
			QueryTimeout: a.configs.Cache.QueryTimeout.Value,
		})
	}

	// Read JWT and symmetric keys
	jwtPrivateKey, jwtPublicKey, symmetricKey, err := a.readAuthKeys()
	if err != nil {
		return err
	}

	// Initialize basic services
	if err := a.initBasicServices(cacheService, symmetricKey); err != nil {
		return err
	}

	// Initialize auth services
	mailService := a.mailServer // Initialize mail service first
	if err := a.initAuthServices(jwtPrivateKey, jwtPublicKey, mailService, cacheService); err != nil {
		return err
	}

	return nil
}

// initBasicServices initializes the core services like health, models, etc.
func (a *App) initBasicServices(cacheService *service.CacheService, symmetricKey []byte) error {
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

	// Resources service
	a.services.Resources, err = service.NewResourcesService(service.ResourcesServiceConf{
		Repository:   a.repositories.Resources,
		CacheService: cacheService,
		OT:           a.telemetry,
	})
	if err != nil {
		return fmt.Errorf("error creating resources service: %w", err)
	}

	// Policies service
	a.services.Policies, err = service.NewPoliciesService(service.PoliciesServiceConf{
		Repository:       a.repositories.Policies,
		ResourcesService: a.services.Resources,
		CacheService:     cacheService,
		OT:               a.telemetry,
	})
	if err != nil {
		return fmt.Errorf("error creating policies service: %w", err)
	}

	// Roles service
	a.services.Roles, err = service.NewRolesService(service.RolesServiceConf{
		Repository:   a.repositories.Roles,
		CacheService: cacheService,
		OT:           a.telemetry,
	})
	if err != nil {
		return fmt.Errorf("error creating roles service: %w", err)
	}

	// Products service
	a.services.Products, err = service.NewProductsService(service.ProductsServiceConf{
		Repository: a.repositories.Products,
		OT:         a.telemetry,
	})
	if err != nil {
		return fmt.Errorf("error creating products service: %w", err)
	}

	return nil
}

// initAuthServices initializes the authentication and authorization services
func (a *App) initAuthServices(jwtPrivateKey, jwtPublicKey []byte, mailService *mailer.MailService, cacheService *service.CacheService) error {
	var err error

	// Authz service
	a.services.Authz, err = service.NewAuthzService(service.AuthzServiceConf{
		Repository:   a.repositories.Users,
		CacheService: cacheService,
		OT:           a.telemetry,
		RegoQuery:    opa.RegoQuery,
		RegoPolicy:   opa.RegoPolicy,
	})
	if err != nil {
		return fmt.Errorf("error creating authz service: %w", err)
	}

	// Authn service
	a.services.Authn, err = service.NewAuthnService(service.AuthnServiceConf{
		Repository:                  a.repositories.Users,
		MailQueueService:            mailService,
		PrivateKey:                  jwtPrivateKey,
		PublicKey:                   jwtPublicKey,
		Issuer:                      a.configs.Authn.Issuer.Value,
		SenderName:                  a.configs.Mail.SenderName.Value,
		SenderEmail:                 a.configs.Mail.SenderAddress.Value,
		AccessTokenDuration:         a.configs.Authn.AccessTokenDuration.Value,
		RefreshTokenDuration:        a.configs.Authn.RefreshTokenDuration.Value,
		UserVerificationAPIEndpoint: a.configs.Authn.UserVerificationAPIEndpoint.Value,
		UserVerificationTokenTTL:    a.configs.Authn.UserVerificationTokenTTL.Value,
		OT:                          a.telemetry,
	})
	if err != nil {
		return fmt.Errorf("error creating authn service: %w", err)
	}

	// Projects service
	a.services.Projects, err = service.NewProjectsService(service.ProjectsServiceConf{
		Repository: a.repositories.Projects,
		OT:         a.telemetry,
	})
	if err != nil {
		return fmt.Errorf("error creating projects service: %w", err)
	}

	return nil
}

// initCacheClient initializes the cache client based on configuration
func (a *App) initCacheClient() (valkey.Client, error) {
	switch a.configs.Cache.Kind.Value {
	case "valkey":
		valkeyConfig := valkey.ClientOption{
			InitAddress: a.configs.Cache.Addresses.Value,
			Username:    a.configs.Cache.Username.Value,
			Password:    a.configs.Cache.Password.Value,
			SelectDB:    a.configs.Cache.DB.Value,
			ClientName:  appName,
		}

		return valkey.NewClient(valkeyConfig)
	default:
		return nil, nil
	}
}

// readAuthKeys reads the JWT and symmetric keys from files
func (a *App) readAuthKeys() ([]byte, []byte, []byte, error) {
	// Read JWT private key
	jwtPrivateKey, err := os.ReadFile(a.configs.Authn.PrivateKeyFile.Value.Name())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error reading JWT private key file: %w", err)
	}

	// Read JWT public key
	jwtPublicKey, err := os.ReadFile(a.configs.Authn.PublicKeyFile.Value.Name())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error reading JWT public key file: %w", err)
	}

	// Read symmetric key
	symmetricHexKey, err := os.ReadFile(a.configs.Authn.SymmetricKeyFile.Value.Name())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error reading symmetric key file: %w", err)
	}

	// Process symmetric key
	symmetricHexKeyCleaned := strings.TrimRight(string(symmetricHexKey), "\n\r")
	symmetricKey, err := hex.DecodeString(symmetricHexKeyCleaned)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error decoding symmetric key: %w", err)
	}

	return jwtPrivateKey, jwtPublicKey, symmetricKey, nil
}
