package domain

import (
	"errors"
	"testing"
)

// An address is checked more than once on its way in: by the handler, on the
// request's body, and again by the use-case, on the domain input built from
// what the first check returned. The two must agree, so what ValidateEmail
// returns is an address it accepts, and returns unchanged.
//
// net/mail drops the quotes of a quoted local part: "a b"@example.com came
// out as a b@example.com, which the second check refused.
func TestAQuotedLocalPartThatDoesNotSurviveItsQuotesIsRefused(t *testing.T) {
	t.Parallel()

	for _, in := range []string{`"a b"@example.com`, `"a,b"@example.com`, `"a(b)"@example.com`} {
		t.Run(in, func(t *testing.T) {
			t.Parallel()

			out, err := ValidateEmail(in, FieldEmail)
			if err == nil {
				t.Fatalf("ValidateEmail(%q) = %q, nil; want a refusal", in, out)
			}

			if v, ok := errors.AsType[*ValidationError](err); !ok || v.Code != "INVALID_FORMAT" {
				t.Errorf("ValidateEmail(%q) error = %v, want an INVALID_FORMAT ValidationError", in, err)
			}
		})
	}
}

func TestANormalisedAddressValidatesToItself(t *testing.T) {
	t.Parallel()

	for _, in := range []string{
		"ada@example.com", "Ada.Lovelace@Example.com", "<ada@example.com>", "Ada <ada@example.com>",
		`"ab"@example.com`, "a+tag@example.co.uk",
	} {
		t.Run(in, func(t *testing.T) {
			t.Parallel()

			once, err := ValidateEmail(in, FieldEmail)
			if err != nil {
				t.Fatalf("ValidateEmail(%q): %v", in, err)
			}

			twice, err := ValidateEmail(once, FieldEmail)
			if err != nil || twice != once {
				t.Errorf("ValidateEmail(%q) = %q, and again = %q, %v; want the same address", in, once, twice, err)
			}
		})
	}
}

// TestTheLengthIsTheAddressNotTheInput: a display name or a pair of quotes makes the
// input long enough and the address it names too short. The use-case's
// second check of the length refused it.
func TestTheLengthIsTheAddressNotTheInput(t *testing.T) {
	t.Parallel()

	for _, in := range []string{"Ada <a@b.c>", `"00"@0`} {
		t.Run(in, func(t *testing.T) {
			t.Parallel()

			out, err := ValidateEmail(in, FieldEmail)
			if v, ok := errors.AsType[*ValidationError](err); !ok || v.Code != "TOO_SHORT" {
				t.Errorf("ValidateEmail(%q) = %q, %v; want a TOO_SHORT ValidationError", in, out, err)
			}
		})
	}
}

// TestAnAddressHoldsNoWhiteSpace: net/mail keeps a rune above 0x7f, a
// no-break space included, inside an address, and ValidateEmail trims every
// Unicode space from its input: "<ab@cd.ef\u00a0>" was spelled
// "ab@cd.ef\u00a0", and spelled again "ab@cd.ef". Stored in one spelling and
// looked up in the other, the address was not found.
func TestAnAddressHoldsNoWhiteSpace(t *testing.T) {
	t.Parallel()

	for _, in := range []string{"<ab@cd.ef\u00a0>", "\"\u00a0ab\"@cd.ef", "<\u3000ab@cd.ef>", "a\u00a0b@cd.ef"} {
		t.Run(in, func(t *testing.T) {
			t.Parallel()

			out, err := ValidateEmail(in, FieldEmail)
			if v, ok := errors.AsType[*ValidationError](err); !ok || v.Code != "INVALID_FORMAT" {
				t.Errorf("ValidateEmail(%q) = %q, %v; want an INVALID_FORMAT ValidationError", in, out, err)
			}
		})
	}
}

// FuzzValidateEmail holds the property for any input: an address the
// validator returns is one it accepts again, unchanged.
func FuzzValidateEmail(f *testing.F) {
	for _, seed := range []string{
		"ada@example.com", `"a b"@example.com`, `"a@b"@example.com`, `"ab"@example.com`,
		"Ada <ada@example.com>", "a..b@example.com", "@", "", `"\"@example.com`, "<ab@cd.ef\u00a0>",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, in string) {
		once, err := ValidateEmail(in, FieldEmail)
		if err != nil {
			return
		}

		twice, err := ValidateEmail(once, FieldEmail)
		if err != nil || twice != once {
			t.Fatalf("ValidateEmail(%q) = %q, and again = %q, %v", in, once, twice, err)
		}
	})
}
