package domain

import (
	"strings"
	"testing"

	"uuid"
)

// A project that is not the caller's is refused with the id the caller sent,
// and in the same sentence when the write found a project gone and cannot
// say which of the ids it was. The handlers answer with the error's text.
func TestAProjectThatIsNotTheCallersSaysSo(t *testing.T) {
	t.Parallel()

	id := uuid.NewV7()

	if got, want := (&ProjectNotOfCallerError{ProjectID: id}).Error(), "invalid project_ids: "+id.String()+" is not one of your projects"; got != want {
		t.Errorf("with the id: %q, want %q", got, want)
	}

	got := (&ProjectNotOfCallerError{}).Error()
	if want := "invalid project_ids: one of the ids is not one of your projects"; got != want {
		t.Errorf("without an id: %q, want %q", got, want)
	}

	if strings.Contains(got, uuid.Nil().String()) {
		t.Errorf("the refusal names the nil id: %q", got)
	}
}
