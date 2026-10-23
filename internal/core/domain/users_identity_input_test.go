package domain

import (
	"testing"

	"uuid"
)

// An account provisioned from a provider identity is stored with it,
// and the identity is the account's: one for another id is refused before
// anything is written, and so is one that is not an identity.
func TestAnAccountsIdentityIsTheAccounts(t *testing.T) {
	t.Parallel()

	id, idp := uuid.NewV7(), uuid.NewV7()

	account := func(identity *LinkUserIdentityInput) *InsertUserInput {
		return &InsertUserInput{
			ID: id, FirstName: "Jane", LastName: "Doe", Email: "sso@example.com",
			Password: "Quartz9Falcon!", LocalAccount: new(false), Identity: identity,
		}
	}

	if err := account(nil).Validate(); err != nil {
		t.Fatalf("an account with no identity: %v", err)
	}

	if err := account(&LinkUserIdentityInput{UserID: id, IDPID: idp, Subject: "sub-1", Email: "sso@example.com"}).Validate(); err != nil {
		t.Errorf("an account with its own identity: %v", err)
	}

	if err := account(&LinkUserIdentityInput{UserID: uuid.NewV7(), IDPID: idp, Subject: "sub-1", Email: "sso@example.com"}).Validate(); err == nil {
		t.Error("an identity for another account was accepted")
	}

	if err := account(&LinkUserIdentityInput{UserID: id, IDPID: idp}).Validate(); err == nil {
		t.Error("an identity with no subject was accepted")
	}

	local := account(&LinkUserIdentityInput{UserID: id, IDPID: idp, Subject: "sub-1", Email: "sso@example.com"})
	local.LocalAccount = new(true)

	if err := local.Validate(); err == nil {
		t.Error("a local account was provisioned with a provider identity")
	}
}
