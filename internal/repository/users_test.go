package repository

import (
	"testing"
	"time"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/model"
	"github.com/stretchr/testify/assert"
)

// TestNewUsersRepository_InvalidConfig tests the NewUsersRepository function with invalid configurations
func TestNewUsersRepository_InvalidConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  UsersRepositoryConfig
		wantErr bool
		errType error
	}{
		{
			name: "Nil DB",
			config: UsersRepositoryConfig{
				DB:              nil,
				MaxPingTimeout:  100 * time.Millisecond,
				MaxQueryTimeout: 100 * time.Millisecond,
				OT:              nil,
				MetricsPrefix:   "test",
			},
			wantErr: true,
			errType: &model.InvalidDBConfigurationError{Message: "invalid database configuration. It is nil"},
		},
		{
			name: "Invalid ping timeout",
			config: UsersRepositoryConfig{
				DB:              nil,
				MaxPingTimeout:  5 * time.Millisecond,
				MaxQueryTimeout: 100 * time.Millisecond,
				OT:              nil,
				MetricsPrefix:   "test",
			},
			wantErr: true,
			errType: &model.InvalidDBConfigurationError{Message: "invalid database configuration. It is nil"}, // DB error takes precedence
		},
		{
			name: "Invalid query timeout",
			config: UsersRepositoryConfig{
				DB:              nil,
				MaxPingTimeout:  100 * time.Millisecond,
				MaxQueryTimeout: 5 * time.Millisecond,
				OT:              nil,
				MetricsPrefix:   "test",
			},
			wantErr: true,
			errType: &model.InvalidDBConfigurationError{Message: "invalid database configuration. It is nil"}, // DB error takes precedence
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := NewUsersRepository(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.Equal(t, tt.errType, err)
				}
				assert.Nil(t, repo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, repo)
			}
		})
	}
}
