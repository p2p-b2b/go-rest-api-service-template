# Description of the project

This is an REST API template in go language. It allows basic CRUD operations over users entities.

The REST API is built using go and the standard library, and it is designed to be simple and easy to use.
The API is also designed to be extensible, allowing for future enhancements and features to be added easily.

## Project Best Practices

This project follows the best practices of Go programming language, including the [GO Style Guide](https://google.github.io/styleguide/go/) and
its components:

- [The Style Guide](https://google.github.io/styleguide/go/guide): outlines the foundation of Go style at Google. This document is definitive and is used as the basis for the recommendations in Style Decisions and Best Practices.
- [Style Decisions](https://google.github.io/styleguide/go/decisions): is a more verbose document that summarizes decisions on specific style points and discusses the reasoning behind the decisions where appropriate.
- [Best Practices](https://google.github.io/styleguide/go/best-practices): is a more verbose document that summarizes decisions on specific style points and discusses the reasoning behind the decisions where appropriate.
- [Effective Go](https://golang.org/doc/effective_go.html): documents some of the patterns that have evolved over time that solve common problems, read well, and are robust to code maintenance needs.

## The project structure

```text
├── cmd -> The folder with the entry points of the applications, this is part of the best practices of Go
│   ├── go-rest-api-service-template
│   │   └── main.go -> The entry point of the REST API service
├── Containerfile -> The file with the definition of the container image
├── database
│   ├── migrations
│   │   └── <numerated migrations files with .sql extension>
│   ├── migrations.go -> The code apply the migrations to the database using goose go library
│   └── README.md -> The README file with the instructions to run the migrations
├── dev-env
│   ├── configuration
│   │   ├── grafana
│   │   │   ├── dashboard
│   │   │   │   ├── default.yaml -> Grafana dashboard configuration file
│   │   │   │   ├── microservices.json -> Grafana dashboard configuration files
│   │   │   │   └── microservicesGroup.json -> Grafana dashboard configuration files
│   │   │   └── datasource
│   │   │       └── grafana-ds.yaml -> Grafana datasource configuration file
│   │   ├── prometheus
│   │   │   └── prometheus.yaml -> Prometheus configuration file
│   │   └── tempo
│   │       └── tempo-local-config.yaml -> Tempo configuration file
│   └── provisioning
│       └── dev-service-pod.yaml -> Podman pod definition file with all the services needed to run the application in local for development
├── dev.env -> The file with the environment variables needed to run the application in local for development
├── docs
│   ├── docs.go -> The file with the documentation of the API in Go format generated automatically when project is built
│   ├── markdown.md -> The file with the documentation of the API in markdown format generated automatically when project is built
│   ├── swagger.json -> The file containing the API documentation in JSON format generated automatically when project is built
│   └── swagger.yaml -> The file containing the API documentation in Swagger format generated automatically when project is built
├── go.mod -> The file with the dependencies of the project
├── go.sum -> The file with the dependencies of the project
├── internal -> The folder with the internal packages of the project, this is part of the best practices of Go
│   ├── config -> The folder with the configuration wrapper of the project used in the flags and env variables settings
│   │   └── <list of go files with the configuration of the project and used in the flags and env variables settings>
│   ├── app -> The folder with the application logic of the project
│   │   ├── app.go -> The file with the application logic of the project
│   ├── http -> The http package with the implementation of the REST API
│   │   ├── handler -> The package handlers of the REST API
│   │   │   └── <list of go files with the handlers of the REST API>
│   │   ├── middleware -> The package middlewares of the REST API
│   │   │   ├── middlewares.go -> The file with the middlewares of the REST API
│   │   │   └── writer.go -> The file with the writer of the REST API
│   │   ├── respond -> The package respond of the REST API used to send the responses
│   │   │   └── http.go -> The file with the respond of the REST API
│   │   └── server -> The package server of the REST API
│   │       └── http.go -> The file with the server of the REST API
│   ├── httpretry -> The package with the implementation of the HTTP retry
│   │   └── <list of go files with the implementation of the HTTP retry>
│   ├── model -> The package with the implementation of the models
│   │   └── <list of go files with the implementation of the models, this includes the models used in the database and the models used in the REST API>
│   ├── o11y -> The package responsible for observability, including metrics, tracing, and logging
│   │   └── <list of go files with the implementation of the observability>
│   ├── repository -> The package with the implementation of the repository
│   │   └── <list of go files with the implementation of the repository>
│   ├── service -> The package with the implementation of the service
│   │   └── <list of go files with the implementation of the service>
│   ├── version -> The package with the versioning information
│   │   └── version.go -> The file with the versioning implementation
├── Makefile -> The file with the Makefile for the project, this is used to run the commands in the project, like build, test, etc.
├── mocks -> The folder with the mock implementations for testing
│   └── <list of go files with the mock implementations for testing>
├── README.md -> The README file with the instructions to run the project
├── test -> The folder containing the test files for the project
│   ├── integration -> The package integration of test of the application, some kind of end to end test for the REST API
│   │   └── <List of go files implementing the integration tests>
```

## The project Software Architecture

The project is built using the Clean Architecture principles, which separates the application into different layers. The main layers are:

- **Model Layer**: This layer contains the definition of the models used in the application. This includes the models used in the database and the models used in the REST API.
- **Repository Layer**: This layer contains the implementation of the repository pattern, which is used to access the database. This layer is responsible for the data access and manipulation.
- **Service Layer**: This layer contains the implementation of the service layer, which is responsible for the business logic of the application. This layer is responsible for the data processing and manipulation.
- **Handler Layer**: This layer contains the implementation of the handlers of the REST API. This layer is responsible for the HTTP requests and responses.
- **Middleware Layer**: This layer contains the implementation of the middlewares of the REST API. This layer is responsible for the HTTP middlewares, like authentication, logging, etc.
- **Configuration Layer**: This layer contains the implementation of the configuration of the application. This layer is responsible for the configuration of the application, like the environment variables, flags, etc.
- **Observability Layer**: This layer contains the implementation of the observability of the application. This layer is responsible for the metrics, tracing, and logging of the application.
- **Testing Layer**: This layer contains the implementation of the testing of the application. This layer is responsible for the unit tests, integration tests, and end-to-end tests of the application.

To improve the testing the application implements [Dependency Injection](https://en.wikipedia.org/wiki/Dependency_injection) using the [Uber mockgen](https://pkg.go.dev/go.uber.org/mock/mockgen) and the `go generate` command. This allows to generate the mocks for the interfaces used in the application, making it easier to test the application.

The project also implements the [Repository Pattern](https://en.wikipedia.org/wiki/Repository_pattern) to separate the data access from the business logic. This allows to change the data access layer without affecting the business logic of the application.

## Project Dependencies

- **Go**: The project is built using Go programming language. The project is using Go modules for dependency management.
- **PostgreSQL**: The project is using PostgreSQL as the database. The project is using the `pgx` library for the database access.
- **PGVector**: The project is using PGVector for the vector embeddings. The project is using PGVector for the vector embeddings.
- **jackc/pgx/v5**: The project is using jackc/pgx/v5 for the database access. The project is using jackc/pgx/v5 for the database access.
- **Goose**: The project is using Goose for the database migrations. The project is using Goose for the database migrations.
- **Swagger**: The project is using Swagger for the API documentation. The project is using Swagger for the API documentation.
- **Prometheus**: The project is using Prometheus for the metrics. The project is using Prometheus for the metrics.
- **Grafana**: The project is using Grafana for the monitoring. The project is using Grafana for the monitoring.
- **Tempo**: The project is using Tempo for the tracing. The project is using Tempo for the tracing.
- **Podman**: The project is using Podman for the containerization. The project is using Podman for the containerization.
- **Air** : The project is using Air for the hot reloading. The project is using Air for the hot reloading.
- **Testify**: The project is using Testify for the testing. The project is using Testify for the testing.

## Testing

The project is using the [Testify](https://pkg.go.dev/github.com/stretchr/testify) library for testing. But in most of the case the project is using the Go testing library. The project is using the `go test -race` command to run the tests. The project is using the `go test -v -race` command to run the tests with verbose output.

### Best Practices

- Never use credentials in the code, always use environment variables or flags to pass the credentials to the application.

## Run the application

This application uses [Air](https://github.com/cosmtrek/air) for hot reloading.

```bash
air
```

## Run the tests

## unit tests
To run the tests, use the following command:

```bash
go test -v -race -coverprofile=./build/coverage.txt -covermode=atomic -tags=unit ./...
```

## integration tests

To run the integration tests, use the following command:

```bash
go test -v -race -tags=integration ./test/integration
```