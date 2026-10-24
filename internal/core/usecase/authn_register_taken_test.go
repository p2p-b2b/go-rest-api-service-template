//go:build unit

package usecase

import (
	"errors"
	"testing"

	"uuid"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
)

// TestATakenIDIsNotATakenAddress: the public registration route answers an
// address that already has an account exactly as it answers a new one, and
// mails the owner, so that the route is not a way to learn who has an
// account. That is a rule about the ADDRESS. The route also takes an optional
// id, which the same create refuses when it is taken (the repository's
// conflict on the primary key, with no Email), and that was answered
// "registered" too, with "you already have an account" mailed to an address
// that had none. It is the id's conflict, returned for the handler's 409.
func TestATakenIDIsNotATakenAddress(t *testing.T) {
	t.Parallel()

	id := uuid.NewV7()
	f := newVerifiedFixture(t)
	f.users.createErr = &domain.UserAlreadyExistsError{ID: id}

	err := f.svc.RegisterUser(t.Context(), &domain.RegisterUserInput{
		ID: id, Email: "new@example.com", FirstName: "Jane", LastName: "Doe",
		Password: "Meadow7Lark!", Disabled: new(true), RegisterMethod: domain.RegisterMethodPassword,
	})

	taken, ok := errors.AsType[*domain.UserAlreadyExistsError](err)
	if !ok || taken.Email != "" || taken.ID != id {
		t.Fatalf("got %T (%v), want the id's conflict", err, err)
	}

	if len(f.notifier.exists) != 0 || len(f.notifier.verifications) != 0 {
		t.Errorf("mail was sent to an address that has no account: exists %v, verifications %v", f.notifier.exists, f.notifier.verifications)
	}
}

// TestATakenIDAndATakenAddressIsTheIDsConflict: when both are taken, the
// database reports whichever unique index it checks first, which depends on
// how the table was built. Answered by that, a taken id (the caller's to
// choose) would tell a taken address from a free one: "registered" for one,
// 409 for the other. The id's conflict is the answer either way.
func TestATakenIDAndATakenAddressIsTheIDsConflict(t *testing.T) {
	t.Parallel()

	id := uuid.NewV7()
	f := newVerifiedFixture(t)
	f.users.add(&domain.User{ID: id, Email: "taken@example.com"})
	f.users.createErr = &domain.UserAlreadyExistsError{Email: "taken@example.com"}

	err := f.svc.RegisterUser(t.Context(), &domain.RegisterUserInput{
		ID: id, Email: "taken@example.com", FirstName: "Jane", LastName: "Doe",
		Password: "Meadow7Lark!", Disabled: new(true), RegisterMethod: domain.RegisterMethodPassword,
	})

	taken, ok := errors.AsType[*domain.UserAlreadyExistsError](err)
	if !ok || taken.Email != "" || taken.ID != id {
		t.Fatalf("got %T (%v), want the id's conflict", err, err)
	}

	if len(f.notifier.exists) != 0 {
		t.Errorf("the address was told apart through a taken id: mailed %v", f.notifier.exists)
	}
}

// TestARegistrationNamesItsMethod: the method decides whether the account is
// a local one, whether its address still has to be proven and how a taken
// address is answered. An unknown or missing one created a verified account
// with no password and no mail.
func TestARegistrationNamesItsMethod(t *testing.T) {
	t.Parallel()

	for _, method := range []domain.RegisterMethod{"", "magic"} {
		t.Run("method_"+string(method), func(t *testing.T) {
			t.Parallel()

			f := newVerifiedFixture(t)

			err := f.svc.RegisterUser(t.Context(), &domain.RegisterUserInput{
				ID: uuid.NewV7(), Email: "new@example.com", FirstName: "Jane", LastName: "Doe",
				Password: "Meadow7Lark!", RegisterMethod: method,
			})

			if _, ok := errors.AsType[*domain.ValidationErrors](err); !ok {
				t.Fatalf("got %T (%v), want a validation error", err, err)
			}

			if len(f.users.created) != 0 {
				t.Error("an account was created")
			}
		})
	}
}
