//go:build unit

package domain

import (
	"strings"
	"testing"
)

// TestAFailuresDetailIsNeverInItsMessage: the detail is the operator's, and
// an error's Error() is what a handler may write into a body. Every type
// that carries one keeps it out of its message.
func TestAFailuresDetailIsNeverInItsMessage(t *testing.T) {
	t.Parallel()

	const detail = "the provider's own words: invalid_client"

	for name, err := range map[string]interface {
		error
		FailureDetail() string
	}{
		"IDPUnreachableError":       &IDPUnreachableError{Message: "the identity provider is not reachable", Detail: detail},
		"IDPSignInRefusedError":     &IDPSignInRefusedError{Message: "the identity provider did not accept this sign-in", Detail: detail},
		"InvalidAuthnServiceError":  &InvalidAuthnServiceError{Message: IDPNotConfiguredMessage, Detail: detail},
		"IDPIdentityNotLinkedError": &IDPIdentityNotLinkedError{Detail: detail},
		"InvalidJWTError":           &InvalidJWTError{Message: "invalid token", Detail: detail},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err.FailureDetail() != detail {
				t.Errorf("FailureDetail() = %q, want the detail", err.FailureDetail())
			}

			if strings.Contains(err.Error(), "own words") {
				t.Errorf("Error() prints the detail: %q", err.Error())
			}
		})
	}
}
