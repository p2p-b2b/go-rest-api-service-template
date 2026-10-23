package o11y

import "sync/atomic"

// AttrOutcome says how a call at a layer ended: one of the Outcome*
// constants, on app_calls_total and app_call_duration_seconds.
const AttrOutcome = "app.outcome"

// How a call at a layer ended.
const (
	// OutcomeOK: it returned no error.
	OutcomeOK = "ok"

	// OutcomeRefused: it returned an error the caller of the service caused
	// -- a bad field, a row that is not there, a wrong password, a limit --
	// and the service answers by saying so. The service worked.
	OutcomeRefused = "refused"

	// OutcomeFailed: it returned any other error. The service, or something
	// it depends on, did not work.
	OutcomeFailed = "failed"
)

// refusal holds the classifier, set once by the composition root.
var refusal atomic.Pointer[func(error) bool]

// SetRefusalClassifier tells the recorder which errors are refusals.
//
// # Why a refusal is not a failure, in a metric as in a log
//
// `successful="false"` counted both. A burst of bad requests -- a client in
// a retry loop on a 404, a scanner, a form that validates on the server --
// raised every error ratio on every dashboard, and a database that had
// really gone away raised it no more than that. The one number an operator
// reads to ask "is the service broken?" could not answer it. The log had
// the same fault, and for the same reason: nothing told a refusal from a
// failure.
//
// # Why the classifier comes from outside
//
// What makes an error a refusal is how the service answers it: with a 4xx.
// That knowledge is one table, in the HTTP adapter (respond.StatusFor), and
// this package cannot import it -- respond imports this one. So the
// composition root hands the question over as a function, and the table
// stays the only list: an error type added to it is classified here without
// a second place to remember.
//
// # Why it is decided per call, and not when the request ends
//
// The log line waits for the HTTP status because it is one line per request.
// A metric here is one count per call at a layer, recorded when the call
// returns; it cannot wait, and it should not. A first sign-in through an
// identity provider looks the identity up and is told "not found": that
// repository call was refused, whatever the request goes on to do. The
// use-case recovers from it and its own call is ok. Each call says how it
// ended, and the request's own outcome is the handler layer's.
//
// Unset -- a test, a tool that never serves a request -- every error is a
// failure, which is the careful answer.
func SetRefusalClassifier(isRefusal func(error) bool) {
	if isRefusal == nil {
		refusal.Store(nil)

		return
	}

	refusal.Store(&isRefusal)
}

// Outcome says how a call that returned err ended.
func Outcome(err error) string {
	if err == nil {
		return OutcomeOK
	}

	if isRefusal := refusal.Load(); isRefusal != nil && (*isRefusal)(err) {
		return OutcomeRefused
	}

	return OutcomeFailed
}
