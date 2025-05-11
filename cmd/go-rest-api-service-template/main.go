package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/app"
)

// main is the entry point of the application
//
//	@title			Go REST API Service Template
//	@version		v1
//	@contact.name	API Support
//	@contact.url	https://qu3ry.me
//	@contact.email	info@qu3ry.me
//	@description	This is a service template for building RESTful APIs in Go.
//	@description	It uses a PostgreSQL database to store user information.
//	@description	The service provides:
//	@description	- CRUD operations for users.
//	@description	- Health and version endpoints.
//	@description	- Configuration using environment variables or command line arguments.
//	@description	- Debug mode to enable debug logging.
//	@description	- TLS enabled to secure the communication.
//
// main function initializes the application, sets up the database connection, and starts the HTTP server
func main() {
	ctx := context.Background()

	application, err := app.NewApp(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing application: %v\n", err)
		os.Exit(1)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
