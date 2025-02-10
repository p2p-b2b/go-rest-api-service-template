package config

import (
	"os"
	"testing"
)

func TestNewDatabaseConfig(t *testing.T) {
	config := NewDatabaseConfig()

	if config.Kind.Value != DefaultDatabaseKind {
		t.Errorf("Expected Kind to be %s, got %s", DefaultDatabaseKind, config.Kind.Value)
	}
	if config.Address.Value != DefaultDatabaseAddress {
		t.Errorf("Expected Address to be %s, got %s", DefaultDatabaseAddress, config.Address.Value)
	}
	if config.Port.Value != DefaultDatabasePort {
		t.Errorf("Expected Port to be %d, got %d", DefaultDatabasePort, config.Port.Value)
	}
	if config.Username.Value != DefaultDatabaseUsername {
		t.Errorf("Expected Username to be %s, got %s", DefaultDatabaseUsername, config.Username.Value)
	}
	if config.Password.Value != DefaultDatabasePassword {
		t.Errorf("Expected Password to be %s, got %s", DefaultDatabasePassword, config.Password.Value)
	}
	if config.Name.Value != DefaultDatabaseName {
		t.Errorf("Expected Name to be %s, got %s", DefaultDatabaseName, config.Name.Value)
	}
	if config.SSLMode.Value != DefaultDatabaseSSLMode {
		t.Errorf("Expected SSLMode to be %s, got %s", DefaultDatabaseSSLMode, config.SSLMode.Value)
	}
	if config.TimeZone.Value != DefaultDatabaseTimeZone {
		t.Errorf("Expected TimeZone to be %s, got %s", DefaultDatabaseTimeZone, config.TimeZone.Value)
	}
}

func TestParseEnvVars_database(t *testing.T) {
	os.Setenv("DATABASE_KIND", "postgres")
	os.Setenv("DATABASE_ADDRESS", "127.0.0.1")
	os.Setenv("DATABASE_PORT", "5433")
	os.Setenv("DATABASE_USERNAME", "testuser")
	os.Setenv("DATABASE_PASSWORD", "testpass")
	os.Setenv("DATABASE_NAME", "testdb")
	os.Setenv("DATABASE_SSL_MODE", "require")
	os.Setenv("DATABASE_TIME_ZONE", "PST")

	config := NewDatabaseConfig()
	config.ParseEnvVars()

	if config.Kind.Value != "postgres" {
		t.Errorf("Expected Kind to be postgres, got %s", config.Kind.Value)
	}
	if config.Address.Value != "127.0.0.1" {
		t.Errorf("Expected Address to be 127.0.0.1, got %s", config.Address.Value)
	}
	if config.Port.Value != 5433 {
		t.Errorf("Expected Port to be 5433, got %d", config.Port.Value)
	}
	if config.Username.Value != "testuser" {
		t.Errorf("Expected Username to be testuser, got %s", config.Username.Value)
	}
	if config.Password.Value != "testpass" {
		t.Errorf("Expected Password to be testpass, got %s", config.Password.Value)
	}
	if config.Name.Value != "testdb" {
		t.Errorf("Expected Name to be testdb, got %s", config.Name.Value)
	}
	if config.SSLMode.Value != "require" {
		t.Errorf("Expected SSLMode to be require, got %s", config.SSLMode.Value)
	}
	if config.TimeZone.Value != "PST" {
		t.Errorf("Expected TimeZone to be PST, got %s", config.TimeZone.Value)
	}

	// Clean up environment variables
	os.Unsetenv("DATABASE_KIND")
	os.Unsetenv("DATABASE_ADDRESS")
	os.Unsetenv("DATABASE_PORT")
	os.Unsetenv("DATABASE_USERNAME")
	os.Unsetenv("DATABASE_PASSWORD")
	os.Unsetenv("DATABASE_NAME")
	os.Unsetenv("DATABASE_SSL_MODE")
	os.Unsetenv("DATABASE_TIME_ZONE")
}

func TestValidate_database(t *testing.T) {
	config := NewDatabaseConfig()

	// Test valid configuration
	err := config.Validate()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test invalid Kind
	config.Kind.Value = "invalid"
	err = config.Validate()
	if err != ErrDatabaseInvalidKind {
		t.Errorf("Expected error %v, got %v", ErrDatabaseInvalidKind, err)
	}
	config.Kind.Value = DefaultDatabaseKind

	// Test invalid Port
	config.Port.Value = -1
	err = config.Validate()
	if err != ErrDatabaseInvalidPort {
		t.Errorf("Expected error %v, got %v", ErrDatabaseInvalidPort, err)
	}
	config.Port.Value = DefaultDatabasePort

	// Test invalid Username
	config.Username.Value = ""
	err = config.Validate()
	if err != ErrDatabaseInvalidUsername {
		t.Errorf("Expected error %v, got %v", ErrDatabaseInvalidUsername, err)
	}
	config.Username.Value = DefaultDatabaseUsername

	// Test invalid Password
	config.Password.Value = ""
	err = config.Validate()
	if err != ErrDatabaseInvalidPassword {
		t.Errorf("Expected error %v, got %v", ErrDatabaseInvalidPassword, err)
	}
	config.Password.Value = DefaultDatabasePassword

	// Test invalid Name
	config.Name.Value = ""
	err = config.Validate()
	if err != ErrDatabaseInvalidDatabaseName {
		t.Errorf("Expected error %v, got %v", ErrDatabaseInvalidDatabaseName, err)
	}
	config.Name.Value = DefaultDatabaseName

	// Test invalid SSLMode
	config.SSLMode.Value = "invalid"
	err = config.Validate()
	if err != ErrDatabaseInvalidSSLMode {
		t.Errorf("Expected error %v, got %v", ErrDatabaseInvalidSSLMode, err)
	}
	config.SSLMode.Value = DefaultDatabaseSSLMode

	// Test invalid TimeZone
	config.TimeZone.Value = ""
	err = config.Validate()
	if err != ErrDatabaseInvalidTimeZone {
		t.Errorf("Expected error %v, got %v", ErrDatabaseInvalidTimeZone, err)
	}
	config.TimeZone.Value = DefaultDatabaseTimeZone

	// Test invalid MaxIdleConns
	config.MaxIdleConns.Value = -1
	err = config.Validate()
	if err != ErrDatabaseInvalidMaxIdleConns {
		t.Errorf("Expected error %v, got %v", ErrDatabaseInvalidMaxIdleConns, err)
	}
	config.MaxIdleConns.Value = DefaultDatabaseMaxIdleConns

	// Test invalid MaxOpenConns
	config.MaxOpenConns.Value = -1
	err = config.Validate()
	if err != ErrDatabaseInvalidMaxOpenConns {
		t.Errorf("Expected error %v, got %v", ErrDatabaseInvalidMaxOpenConns, err)
	}
	config.MaxOpenConns.Value = DefaultDatabaseMaxOpenConns

	// Test invalid MaxPingTimeout
	config.MaxPingTimeout.Value = 0
	err = config.Validate()
	if err != ErrDatabaseInvalidMaxPingTimeout {
		t.Errorf("Expected error %v, got %v", ErrDatabaseInvalidMaxPingTimeout, err)
	}
	config.MaxPingTimeout.Value = DefaultDatabaseMaxPingTimeout

	// Test invalid MaxQueryTimeout
	config.MaxQueryTimeout.Value = 0
	err = config.Validate()
	if err != ErrDatabaseInvalidMaxQueryTimeout {
		t.Errorf("Expected error %v, got %v", ErrDatabaseInvalidMaxQueryTimeout, err)
	}
	config.MaxQueryTimeout.Value = DefaultDatabaseMaxQueryTimeout

	// Test invalid ConnMaxIdleTime
	config.ConnMaxIdleTime.Value = 0
	err = config.Validate()
	if err != ErrDatabaseInvalidConnMaxIdleTime {
		t.Errorf("Expected error %v, got %v", ErrDatabaseInvalidConnMaxIdleTime, err)
	}
	config.ConnMaxIdleTime.Value = DefaultDatabaseConnMaxIdleTime

	// Test invalid ConnMaxLifetime
	config.ConnMaxLifetime.Value = 0
	err = config.Validate()
	if err != ErrDatabaseInvalidConnMaxLifetime {
		t.Errorf("Expected error %v, got %v", ErrDatabaseInvalidConnMaxLifetime, err)
	}
	config.ConnMaxLifetime.Value = DefaultDatabaseConnMaxLifetime
}
