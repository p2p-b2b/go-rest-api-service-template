//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/config"
)

// testDBPool is a shared connection pool for all integration tests
var testDBPool *pgxpool.Pool

const (
	// API endpoint for the test suite
	apiEndpointURL = "http://localhost:8080/api/v1"

	// mailServerEndpointURL is the endpoint for email testing
	mailServerEndpointURL = "http://localhost:8025/api/v1"
	verifyEmailAddress    = config.DefaultMailSenderAddress

	// Database connection parameters
	dbHost            = "localhost"
	dbPort            = 5432
	dbSSLMode         = "disable"
	dbName            = "go-rest-api-service-template"
	dbTimeZone        = "UTC"
	dbMaxConns        = 20
	dbMinConns        = 1
	dbMaxConnLifetime = 30 * time.Minute
	dbMaxConnIdleTime = 5 * time.Minute

	dbUsernameEnvVarName = "DB_USERNAME"
	dbPasswordEnvVarName = "DB_PASSWORD"
)

// TestMain is the entry point for the test suite
func TestMain(m *testing.M) {
	setupTestSuite()

	// Run the tests
	code := m.Run()

	// Teardown code can go here if needed
	tearDownTestSuite()

	// Exit with the test result code
	os.Exit(code)
}

func setupTestSuite() {
	fmt.Println("🧪 Setting up integration test environment")

	fmt.Print("⚙️ Setting up environment variables from file...")
	if err := config.SetEnvVarFromFile(); err != nil {
		fmt.Println("❌ Error setting environment variables from file:", err)
		os.Exit(1)
	}
	fmt.Println("✅")

	// Set up the database connection pool
	setupTestDB()

	fmt.Println("🧪 Setting integration test done... ✅")
}

func tearDownTestSuite() {
	fmt.Println("🔨 Tearing down integration test environment...")

	fmt.Print("🛢  Closing database connection pool...")
	testDBPool.Close()
	testDBPool = nil
	fmt.Println("✅")

	fmt.Print("🛢  Unsetting environment variables...")
	os.Unsetenv(dbUsernameEnvVarName)
	os.Unsetenv(dbPasswordEnvVarName)
	fmt.Println("✅")

	fmt.Print("🛢  Deleting all emails from the mail server...")
	if err := deleteAllEmails(); err != nil {
		fmt.Println("❌ Error deleting emails from mail server:", err)
	} else {
		fmt.Println("✅")
	}

	fmt.Println("🔨 Teardown integration test done... ✅")
}

func setupTestDB() {
	fmt.Print("🛢  Getting database user and password from environment variables...")
	dbUser := config.GetEnv(dbUsernameEnvVarName, "")
	dbPassword := config.GetEnv(dbPasswordEnvVarName, "")
	fmt.Println("✅")

	fmt.Print("🛢  Validating database user and password...")
	if dbUser == "" || dbPassword == "" {
		fmt.Println("❌ DB_USERNAME or DB_PASSWORD environment variable is not set or empty")
		os.Exit(1)
	}
	fmt.Println("✅")

	dbDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		dbHost,
		dbPort,
		dbUser,
		dbPassword,
		dbName,
		dbSSLMode,
		dbTimeZone,
	)

	fmt.Print("🛢  Parsing database connection string...")
	config, err := pgxpool.ParseConfig(dbDSN)
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to parse database config: %v", err))
	}
	fmt.Println("✅")

	fmt.Print("🛢  Setting up database connection pool...")
	testDBPool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		panic(fmt.Sprintf("❌  Failed to create database connection pool: %v", err))
	}
	fmt.Println("✅")
}
