package templates

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEmailAccountVerification(t *testing.T) {
	tests := []struct {
		name    string
		conf    *EmailAccountVerificationConf
		wantErr error
		want    *EmailAccountVerification
	}{
		{
			name: "Valid configuration",
			conf: &EmailAccountVerificationConf{
				VerificationAPIEndpoint: "http://localhost:8080/verify",
				VerificationToken:       "testtoken",
				VerificationTTL:         "1h",
				UserName:                "Test User",
				HTML:                    true,
			},
			wantErr: nil,
			want: &EmailAccountVerification{
				verificationAPIEndpoint: "http://localhost:8080/verify",
				verificationToken:       "testtoken",
				verificationTTL:         "1h",
				userName:                "Test User",
				html:                    true,
			},
		},
		{
			name:    "Nil configuration",
			conf:    nil,
			wantErr: ErrInvalidAccountVerificationConf,
			want:    nil,
		},
		{
			name: "Empty VerificationAPIEndpoint",
			conf: &EmailAccountVerificationConf{
				VerificationToken: "testtoken",
				VerificationTTL:   "1h",
				UserName:          "Test User",
			},
			wantErr: ErrInvalidVerificationAPIEndpoint,
			want:    nil,
		},
		{
			name: "Empty VerificationToken",
			conf: &EmailAccountVerificationConf{
				VerificationAPIEndpoint: "http://localhost:8080/verify",
				VerificationTTL:         "1h",
				UserName:                "Test User",
			},
			wantErr: ErrInvalidVerificationToken,
			want:    nil,
		},
		{
			name: "Empty VerificationTTL",
			conf: &EmailAccountVerificationConf{
				VerificationAPIEndpoint: "http://localhost:8080/verify",
				VerificationToken:       "testtoken",
				UserName:                "Test User",
			},
			wantErr: ErrInvalidVerificationTTL,
			want:    nil,
		},
		{
			name: "Empty UserName",
			conf: &EmailAccountVerificationConf{
				VerificationAPIEndpoint: "http://localhost:8080/verify",
				VerificationToken:       "testtoken",
				VerificationTTL:         "1h",
			},
			wantErr: ErrInvalidUserName,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEmailAccountVerification(tt.conf)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEmailAccountVerification_Render(t *testing.T) {
	tests := []struct {
		name       string
		conf       *EmailAccountVerificationConf
		expectHTML bool
	}{
		{
			name: "Render HTML template",
			conf: &EmailAccountVerificationConf{
				VerificationAPIEndpoint: "http://example.com/verify",
				VerificationToken:       "htmltoken",
				VerificationTTL:         "30m",
				UserName:                "HTML User",
				HTML:                    true,
			},
			expectHTML: true,
		},
		{
			name: "Render Text template",
			conf: &EmailAccountVerificationConf{
				VerificationAPIEndpoint: "http://example.com/verify",
				VerificationToken:       "texttoken",
				VerificationTTL:         "15m",
				UserName:                "Text User",
				HTML:                    false,
			},
			expectHTML: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emailVerifier, err := NewEmailAccountVerification(tt.conf)
			assert.NoError(t, err)
			assert.NotNil(t, emailVerifier)

			renderedOutput := emailVerifier.Render()

			// Additional checks based on template type
			if tt.expectHTML {
				assert.Contains(t, renderedOutput, "<!DOCTYPE html>")
				assert.Contains(t, renderedOutput, "<h1>Account Verification</h1>")
				assert.Contains(t, renderedOutput, `<a href="http://example.com/verify/htmltoken">Verify Account</a>`)
				assert.Contains(t, renderedOutput, "Token will expire in 30m")
			} else {
				assert.NotContains(t, renderedOutput, "<!DOCTYPE html>")
				assert.Contains(t, renderedOutput, "Account Verification")
				assert.Contains(t, renderedOutput, "http://example.com/verify/texttoken")
				assert.Contains(t, renderedOutput, "Token will expire in 15m")
			}
		})
	}
}
