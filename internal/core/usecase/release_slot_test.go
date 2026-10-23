//go:build unit

package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
	"github.com/slashdevops/go-rest-api-service-template/internal/o11y"
)

// releaseSeen records the context a release was asked on, and answers err,
// noting it where the layers under the real decrement do.
type releaseSeen struct {
	ResourcesLimitsServiceConsumer

	contextErr error
	deadline   time.Time
	carried    any
	err        error
}

type releaseCarriedKey struct{}

func (r *releaseSeen) DecrementUsage(ctx context.Context, _ domain.ResourcesLimitsScope, _ domain.ResourcesLimitsResourceType) error {
	r.contextErr = ctx.Err()
	r.deadline, _ = ctx.Deadline()
	r.carried = ctx.Value(releaseCarriedKey{})

	if r.err != nil {
		o11y.NoteFailure(ctx, r.err)
	}

	return r.err
}

// A creation that fails gives its reserved slot back, and the request being
// cancelled is one of the ways it fails: a client that walks away. The
// release ran on the request's context and failed the same way, leaving the
// counter one too high. It runs detached now, with an end of its own, and
// keeps what the request carried; and a release that fails is logged, not
// put in the request's failure holder, whose failure is the creation's.
func TestAReleasedSlotIsGivenBackAfterTheRequestIsCancelled(t *testing.T) {
	t.Parallel()

	creation := errors.New("the create failed")
	limits := &releaseSeen{err: errors.New("the counter could not be written")}

	ctx := o11y.WithFailure(context.WithValue(t.Context(), releaseCarriedKey{}, "the request's"))
	o11y.NoteFailure(ctx, creation)

	ctx, cancel := context.WithCancel(ctx)
	cancel()

	ReleaseResourceSlot(ResourceLimitCheckParams{
		Ctx:             ctx,
		ResourcesLimits: limits,
		Scope:           domain.ResourcesLimitsScope{Type: domain.ResourcesLimitsScopeTypeSystem},
		ResourceType:    domain.ResourcesLimitsResourceTypeUsers,
	})

	if limits.contextErr != nil {
		t.Errorf("the release ran on an ended context: %v", limits.contextErr)
	}

	if limits.deadline.IsZero() || time.Until(limits.deadline) > resourceSlotReleaseTimeout {
		t.Errorf("the release has no end of its own: %v", limits.deadline)
	}

	if limits.carried != "the request's" {
		t.Errorf("the release lost what the request carried: %v", limits.carried)
	}

	if failure, _ := o11y.FailureFrom(ctx); !errors.Is(failure.Err, creation) {
		t.Errorf("the request's failure is %v, want the creation's", failure.Err)
	}
}
