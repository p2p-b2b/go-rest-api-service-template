package domain

// FailureDetail is what an operator reads about a failure and a caller never
// does.
//
// Some errors are answered in fixed words on purpose. A sign-in through an
// identity provider is refused in one sentence whatever the reason, so that
// the reason does not tell anyone which addresses have accounts here; a
// provider that cannot be used is one sentence on a public route, so that
// the provider's own text, which a dependency or the provider may change,
// never becomes the API's. The reason was then written as a line of its own
// at WARN, beside the access line that reports the same failure: two lines
// for one failure, at a level that says "act" about a refusal, and with no
// context, so no trace led to them.
//
// The reason travels in the error instead, in a Detail field that Error()
// never prints, and the one line written about the failure carries it as
// `detail` (o11y reads it through this method): the access line of the
// request, at ERROR for a 5xx and in its DEBUG companion for a refusal, or
// `operation_failed` outside a request. error_type is unchanged: the detail
// is text, not a wrapped error, so the bottom of the chain is still the
// domain type.
func (e *IDPUnreachableError) FailureDetail() string { return e.Detail }

// FailureDetail implements the operator's detail; see IDPUnreachableError's.
func (e *IDPSignInRefusedError) FailureDetail() string { return e.Detail }

// FailureDetail implements the operator's detail; see IDPUnreachableError's.
func (e *InvalidAuthnServiceError) FailureDetail() string { return e.Detail }

// FailureDetail implements the operator's detail; see IDPUnreachableError's.
func (e *IDPIdentityNotLinkedError) FailureDetail() string { return e.Detail }

// FailureDetail implements the operator's detail; see IDPUnreachableError's.
func (e *InvalidJWTError) FailureDetail() string { return e.Detail }
