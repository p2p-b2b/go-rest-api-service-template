# go-rest-api-service-template

[![Go Report Card](https://goreportcard.com/badge/github.com/p2p-b2b/go-rest-api-service-template)](https://goreportcard.com/report/github.com/p2p-b2b/go-rest-api-service-template)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/p2p-b2b/go-rest-api-service-template)

This is a comprehensive template for a Go HTTP REST API Service with advanced authentication, authorization, and multi-tenant capabilities.

## Features

### Core Infrastructure

- ✅ Create a new repository from this template
- ✅ Change the module name in `go.mod` using the command `go mod edit -module github.com/your-username/your-repo`
- ✅ Change the service name in `cmd/go-rest-api-service-template` to the name of your service, e.g. `cmd/your-service-name`
- ✅ Change the service name in `Makefile` to the name of your service
- ✅ Use flags, Environment Variables of `.env` file to configure the service
- ✅ Hot reload with [air](https://github.com/cosmtrek/air), use `make install-air` to install it, then `air` to run it
- ✅ Ready to use Certificates for HTTPS, see [Self-Signed Certificates](#self-signed-certificates-for-mutual-tls-mtls)
- ✅ Containerize your service with Podman, see [Building and Running](#building-and-running)
- ✅ Podman pod file for development [dev-service-pod.yaml](dev-env/provisioning/dev-service-pod.yaml), see [Development environment](#development-environment). This is something like `docker-compose/podman-compose` but more powerful
- ✅ Database migrations with [goose](https://github.com/pressly/goose). Check the [database](database/README.md) documentation
- ✅ Validation with [go-playground/validator](https://github.com/go-playground/validator)
- ✅ Allow filtering, sorting, pagination and partial responses
- ✅ Middleware for logging and headers versioning
- ✅ Return JSON as default response, even for standard http library errors

### Authentication & Authorization

- ✅ **JWT-based Authentication** - Secure user authentication with access and refresh tokens
- ✅ **User Registration & Verification** - Email-based user registration with verification workflow
- ✅ **Role-Based Access Control (RBAC)** - Flexible role and permission system
- ✅ **Policy-Based Authorization** - Fine-grained access control with custom policies
- ✅ **Open Policy Agent (OPA) Integration** - Advanced authorization using Rego policies with wildcard support
- ✅ **Multi-tenant Project Isolation** - Projects assigned to users with proper access control
- ✅ **Admin and Regular User Support** - Different permission levels for different user types

### API Features

- ✅ **User Management** - Complete CRUD operations for users with role assignments
- ✅ **Project Management** - Multi-tenant project system with user assignments
- ✅ **Product Management** - Project-scoped product management
- ✅ **Role Management** - Dynamic role creation and assignment
- ✅ **Policy Management** - Custom authorization policies with action/resource matching
- ✅ **Resource Discovery** - Query available resources and permissions

### Security & Performance

- ✅ **JWT Validation** - Robust token validation with configurable expiration
- ✅ **Password Security** - Secure password hashing and validation
- ✅ **Rate Limiting** - Built-in rate limiting capabilities
- ✅ **Input Validation** - Comprehensive request validation
- ✅ **SQL Injection Protection** - Parameterized queries and input sanitization
- ✅ **CORS Support** - Configurable Cross-Origin Resource Sharing

### Observability & Monitoring

- ✅ **OpenTelemetry Integration** - Distributed tracing and metrics
- ✅ **Prometheus Metrics** - Application and business metrics
- ✅ **Grafana Dashboards** - Pre-configured monitoring dashboards
- ✅ **Tempo Tracing** - Distributed tracing backend
- ✅ **Health Checks** - Service health monitoring endpoints
- ✅ **Structured Logging** - JSON-based logging with levels

### Database & Caching

- ✅ **PostgreSQL Integration** - Primary database with advanced features
- ✅ **PGVector Support** - Vector embeddings support for AI/ML applications
- ✅ **Database Migrations** - Versioned schema management with goose
- ✅ **Connection Pooling** - Efficient database connection management
- ✅ **Valkey/Redis Caching** - High-performance caching layer
- ✅ **Transaction Support** - ACID transaction management

## API Endpoints

The service provides comprehensive REST API endpoints for managing users, projects, products, roles, and policies.

### Authentication Endpoints

- `POST /auth/register` - Register a new user
- `POST /auth/login` - Authenticate user and get JWT tokens
- `POST /auth/refresh` - Refresh access token using refresh token
- `POST /auth/logout` - Logout user and invalidate tokens
- `GET /auth/verify/{token}` - Verify user email address

### User Management

- `GET /users` - List all users (with pagination, filtering, sorting)
- `GET /users/{user_id}` - Get specific user details
- `POST /users` - Create a new user (admin only)
- `PUT /users/{user_id}` - Update user information
- `DELETE /users/{user_id}` - Delete a user
- `POST /users/{user_id}/roles` - Assign roles to user
- `DELETE /users/{user_id}/roles` - Remove roles from user
- `GET /users/{user_id}/roles` - List user's roles

### Project Management

- `GET /projects` - List accessible projects
- `GET /projects/{project_id}` - Get project details
- `POST /projects` - Create a new project
- `PUT /projects/{project_id}` - Update project
- `DELETE /projects/{project_id}` - Delete project

### Product Management

- `GET /projects/{project_id}/products` - List products in a project
- `GET /projects/{project_id}/products/{product_id}` - Get product details
- `POST /projects/{project_id}/products` - Create product in project
- `PUT /projects/{project_id}/products/{product_id}` - Update product
- `DELETE /projects/{project_id}/products/{product_id}` - Delete product

### Role & Policy Management

- `GET /roles` - List all roles
- `GET /roles/{role_id}` - Get role details
- `POST /roles` - Create a new role
- `PUT /roles/{role_id}` - Update role
- `DELETE /roles/{role_id}` - Delete role
- `POST /roles/{role_id}/policies` - Link policies to role
- `DELETE /roles/{role_id}/policies` - Unlink policies from role

- `GET /policies` - List all policies
- `GET /policies/{policy_id}` - Get policy details
- `POST /policies` - Create a new policy
- `PUT /policies/{policy_id}` - Update policy
- `DELETE /policies/{policy_id}` - Delete policy

### Resource Discovery

- `GET /resources` - List all available resources
- `GET /resources/{resource_id}` - Get resource details
- `GET /resources/matches` - Find resources matching action/resource criteria

### Health & Monitoring

- `GET /health` - Service health check
- `GET /version` - Service version information
- `GET /metrics` - Prometheus metrics (if enabled)

## Authentication Flow

### User Registration and Verification

1. **Register**: `POST /auth/register` with user details
2. **Email Verification**: User receives verification email
3. **Verify**: Click verification link or use `GET /auth/verify/{token}`
4. **Login**: `POST /auth/login` with credentials to get JWT tokens

### JWT Token Usage

1. **Access Token**: Include in `Authorization: Bearer <token>` header for API calls
2. **Refresh Token**: Use `POST /auth/refresh` when access token expires
3. **Logout**: `POST /auth/logout` to invalidate tokens

### Authorization Model

The service uses a sophisticated authorization model combining:

- **Role-Based Access Control (RBAC)**: Users have roles, roles have policies
- **Policy-Based Authorization**: Fine-grained permissions with action/resource matching
- **Open Policy Agent (OPA)**: Advanced policy evaluation with Rego language
- **Project Isolation**: Multi-tenant architecture with project-level access control

#### Permission Resolution

1. **Admin Users**: Full access to all resources (bypass policy checks)
2. **Regular Users**: Access based on assigned roles and policies
3. **Project Scope**: Users can only access projects they're assigned to
4. **Resource Wildcards**: Support for wildcard patterns in resource matching
5. **Action Permissions**: Granular control over CRUD operations

## Quick Start

### 1. Clone and Setup

```bash
git clone https://github.com/your-username/your-repo.git
cd your-repo

# Update module name
go mod edit -module github.com/your-username/your-repo

# Install dependencies
go mod tidy
```

### 2. Environment Configuration

```bash
# Copy environment template
cp dev.env .env

# Configure your environment variables
# Edit .env file with your database credentials, JWT secrets, etc.
```

### 3. Start Development Environment

```bash
# Start the development stack (PostgreSQL, Redis, etc.)
make start-dev-env

# Run database migrations
make migrate-up

# Start the API server with hot reload
air
```

### 4. Test the API

```bash
# Check service health
curl http://localhost:8080/health

# Register a new user
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@example.com",
    "password": "SecurePassword123!"
  }'

# Login and get tokens
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "SecurePassword123!"
  }'
```

## Configuration

The service supports configuration through:

- **Environment Variables**: Define in `.env` file or system environment
- **Command Line Flags**: Override environment variables
- **Configuration Files**: YAML/JSON configuration support

### Key Configuration Options

```bash
# Server Configuration
HTTP_SERVER_HOST=localhost
HTTP_SERVER_PORT=8080
HTTP_SERVER_READ_TIMEOUT=30s
HTTP_SERVER_WRITE_TIMEOUT=30s

# Database Configuration
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=your_database
DATABASE_USER=your_user
DATABASE_PASSWORD=your_password

# JWT Configuration
JWT_ACCESS_TOKEN_DURATION=15m
JWT_REFRESH_TOKEN_DURATION=7d
JWT_PRIVATE_KEY_PATH=./certs/jwt.key
JWT_PUBLIC_KEY_PATH=./certs/jwt.pub

# Cache Configuration
CACHE_TYPE=valkey
CACHE_HOST=localhost
CACHE_PORT=6379

# Email Configuration
MAIL_SMTP_HOST=localhost
MAIL_SMTP_PORT=1025
MAIL_FROM_ADDRESS=noreply@example.com

# Observability
TELEMETRY_ENABLED=true
TELEMETRY_ENDPOINT=http://localhost:4318
METRICS_ENABLED=true
```

## Testing

The project includes comprehensive unit and integration tests.

### Unit Tests

```bash
# Run unit tests with coverage
go test -v -race -coverprofile=./build/coverage.txt -covermode=atomic -tags=unit ./...
```

### Integration Tests

```bash
# Run integration tests
go test -v -race -tags=integration ./tests/integration -count 1
```

The integration tests require:

- PostgreSQL database
- Redis/Valkey cache
- SMTP server (for email testing)

### Test Environment Setup

The integration tests automatically set up and tear down test environments, including:

- Database schema creation and cleanup
- Test data fixtures
- Email server integration
- Cache layer testing

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for your changes
5. Ensure all tests pass (`make test`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## References

- [RFC-7223 -> Hypertext Transfer Protocol (HTTP/1.1): Semantics and Content](https://www.rfc-editor.org/rfc/rfc7231)
- [RESTful web API design](https://learn.microsoft.com/en-us/azure/architecture/best-practices/api-design)
- [API pagination techniques](https://samu.space/api-pagination/#uuid-primary-keys)
- [How To Do Pagination in Postgres with Golang in 4 Common Ways](https://medium.easyread.co/how-to-do-pagination-in-postgres-with-golang-in-4-common-ways-12365b9fb528)
- ["Cursor Pagination" Profile](https://jsonapi.org/profiles/ethanresnick/cursor-pagination/)
- [Cursor-Based Pagination in Go](https://mtekmir.com/blog/golang-cursor-pagination/)
- [Migrations in YDB using "goose"](https://blog.ydb.tech/migrations-in-ydb-using-goose-58137bc5c303)
- [Multi-hop tracing with OpenTelemetry in Golang](https://faun.pub/multi-hop-tracing-with-opentelemetry-in-golang-792df5feb37c)
- [How to setup your own CA with OpenSSL](https://gist.github.com/soarez/9688998)
- [How to setup your own self signed Certificate Authority (CA) and certificates with OpenSSL 3.x and using ED25519](https://gist.github.com/christiangda/aaa1c5b58dfa17f4d1cf6e42d60f9273#file-howto-self-signed-ca-certs-ed25519-md)
- [Open Policy Agent (OPA)](https://www.openpolicyagent.org/)
- [JWT Best Practices](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-jwt-bcp)

## API Documentation

The Swagger documentation is available at `/swagger/index.html`.

To generate the Swagger documentation, you need to install [swag](https://github.com/swaggo/swag) and run the following command:

```bash
swag init \
  --dir "./cmd/go-rest-api-service-template/,./internal/handler" \
  --output ./docs \
  --parseDependency true \
  --parseInternal true
```

## References

- [RFC-7223 -> Hypertext Transfer Protocol (HTTP/1.1): Semantics and Content](https://www.rfc-editor.org/rfc/rfc7231)
- [RESTful web API design](https://learn.microsoft.com/en-us/azure/architecture/best-practices/api-design)
- [API pagination techniques](https://samu.space/api-pagination/#uuid-primary-keys)
- [How To Do Pagination in Postgres with Golang in 4 Common Ways](https://medium.easyread.co/how-to-do-pagination-in-postgres-with-golang-in-4-common-ways-12365b9fb528)
- [“Cursor Pagination” Profile](https://jsonapi.org/profiles/ethanresnick/cursor-pagination/)
- [Cursor-Based Pagination in Go](https://mtekmir.com/blog/golang-cursor-pagination/)
- [Migrations in YDB using “goose”](https://blog.ydb.tech/migrations-in-ydb-using-goose-58137bc5c303)
- [Multi-hop tracing with OpenTelemetry in Golang](https://faun.pub/multi-hop-tracing-with-opentelemetry-in-golang-792df5feb37c)
- [How to setup your own CA with OpenSSL](https://gist.github.com/soarez/9688998)
- [How to setup your own self signed Certificate Authority (CA) and certificates with OpenSSL 3.x and using ED25519](https://gist.github.com/christiangda/aaa1c5b58dfa17f4d1cf6e42d60f9273#file-howto-self-signed-ca-certs-ed25519-md)

## Swagger Documentation

The Swagger documentation is available at `/swagger/index.html`.

To generate the Swagger documentation, you need to install [swag](https://github.com/swaggo/swag) and run the following command:

```bash
swag init \
  --dir "./cmd/go-rest-api-service-template/,./internal/handler" \
  --output ./docs \
  --parseDependency true \
  --parseInternal true
```

## Building and Running

### Requirement

By default Podman machine (Macbook ARM processors) adds only you $HOME directory to the container.

Reference: <https://github.com/containers/podman/issues/14815>

To add more directories you need to create a new machine with the following command:

```bash
podman machine stop
podman machine rm
podman machine init -v $HOME:$HOME -v /tmp:/tmp
podman machine start
```

**WARNING:** This will remove the current machine and all the containers.

### Development environment

start

```bash
# create the necessary containers
make start-dev-env

# start the development environment
air
```

stop

```bash
# create the necessary containers
make stop-dev-env
```

**OPTIONAL:** Connect to the PostgreSQL database from the host

```bash
PGPASSWORD=password psql -h localhost -p 5432 -U username
```

Database migrations

see [database](database/README.md)

## Project Structure

```text
├── cmd/                              # Application entry points
│   ├── go-rest-api-service-template/ # Main API server
│   ├── apiendpoints/                 # API endpoint documentation tool
│   ├── saltpwd/                      # Password hashing utility
│   └── uuidgen/                      # UUID generation utility
├── internal/                         # Private application code
│   ├── app/                         # Application configuration and setup
│   ├── config/                      # Configuration management
│   ├── http/                        # HTTP layer (handlers, middleware, server)
│   │   ├── handler/                 # Request handlers
│   │   ├── middleware/              # HTTP middleware
│   │   ├── respond/                 # Response utilities
│   │   └── server/                  # HTTP server implementation
│   ├── jwtvalidator/               # JWT validation logic
│   ├── model/                      # Data models and validation
│   ├── o11y/                       # Observability (metrics, tracing)
│   ├── opa/                        # Open Policy Agent integration
│   ├── repository/                 # Data access layer
│   ├── service/                    # Business logic layer
│   ├── templates/                  # Email and other templates
│   └── version/                    # Version information
├── database/                       # Database migrations and documentation
│   └── migrations/                 # SQL migration files
├── tests/                         # Test suites
│   ├── integration/               # Integration tests
│   └── provisioning/              # Test environment setup
├── dev-env/                       # Development environment configuration
│   ├── configuration/             # Service configurations (Grafana, Prometheus)
│   └── provisioning/              # Container orchestration
├── docs/                          # API documentation (Swagger)
├── certs/                         # TLS certificates
└── build/                         # Build artifacts
```

### Key Components

- **Authentication System**: JWT-based with refresh tokens
- **Authorization Engine**: OPA-powered policy evaluation
- **Multi-tenancy**: Project-based resource isolation
- **Observability**: OpenTelemetry integration with Prometheus metrics
- **Database**: PostgreSQL with connection pooling and migrations
- **Caching**: Valkey/Redis for session and authorization data
- **Email**: SMTP integration for user verification
- **Testing**: Comprehensive unit and integration test suites

## Architecture Decisions

### System Authentication Flow

1. **Registration**: Users register with email verification
2. **Login**: JWT access/refresh token pair generation
3. **Authorization**: OPA policy evaluation with cached results
4. **Project Access**: Multi-tenant isolation with user assignments

### Security Model

- **Password Security**: bcrypt hashing with salt
- **JWT Tokens**: RSA256 signing with configurable expiration
- **Authorization**: Policy-based with wildcard resource matching
- **Input Validation**: Comprehensive request validation
- **SQL Safety**: Parameterized queries and input sanitization

### Performance Optimizations

- **Database**: Connection pooling and prepared statements
- **Caching**: Redis-based authorization cache
- **Pagination**: Cursor-based for large datasets
- **Observability**: Efficient telemetry with sampling
