package repositorypg

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/slashdevops/go-rest-api-service-template/internal/core/domain"
)

// The messages the four typed errors every repository answers carry. Our
// words, never pgErr.Message: these reach the client as a 400, and
// Postgres's text is not ours to publish. It named the statement's table
// aliases and quoted the statement ("column usrs.x does not exist"), and a
// Postgres or driver upgrade could reword every one of these answers. The
// failing statement is in the Postgres log; the span keeps the type.
const (
	pgInvalidByteSequence  = "the request contains a byte sequence that is not valid text"
	pgInvalidMessageFormat = "the request could not be read by the database"
	pgUndefinedColumn      = "the request refers to a field that does not exist"
	pgDatatypeMismatch     = "a value has the wrong type for its field"
)

// What a filter says when Postgres would not take it. Our words, never
// Postgres's: its message echoes the caller's value ("invalid input syntax
// for type uuid: \"nope\"") and names the statement's aliases.
const (
	filterValueRefused      = "a value does not fit the type of the field it is compared with"
	filterPatternRefused    = "a pattern is not a valid regular expression or LIKE pattern"
	filterExpressionRefused = "the expression is not one a list can be filtered by"
)

// pgSyntaxError is a statement Postgres could not parse.
const pgSyntaxError = "42601"

// filterRefusals are the SQLSTATEs with which Postgres refuses a filter the
// parser accepted, and what the caller is told.
//
// The parser checks the expression's shape and its field names. It does not
// know the columns' types, so `id = 'nope'`, `disabled = 'maybe'` and
// `id = 1` all pass it, and Postgres refuses the cast or finds no operator.
// Each was a 500 that said nothing, and an ERROR line for a caller's
// mistake.
//
// The set is spelled out, not "class 22": the class also holds what a caller
// cannot cause -- a negative LIMIT (2201W), a parameter the driver encoded
// wrongly (22P03), a division by zero -- and those are ours. A syntax error
// (42601) is not in this table either, because a statement this package
// built wrongly raises it too: it is the filter's only when Postgres points
// inside the filter (see syntaxErrorInFilter).
//
// Two refusals are not here because they were never a 500: a column that
// does not exist (42703) and a text field used as a boolean (42804).
// handlePgError answers both with a 400 and its own fixed sentence. (An
// unquoted word, which Postgres would read as a column, does not get that
// far: the parser refuses it.)
var filterRefusals = map[string]string{
	"22P02": filterValueRefused,   // invalid text representation: 'nope' as a uuid, a boolean, an integer, an enum
	"22007": filterValueRefused,   // invalid datetime format
	"22008": filterValueRefused,   // datetime field overflow: February 30th
	"22009": filterValueRefused,   // invalid time zone displacement
	"22023": filterValueRefused,   // invalid parameter value: a time zone that does not exist
	"22003": filterValueRefused,   // numeric value out of range
	"2202E": filterValueRefused,   // array subscript error
	"54000": filterValueRefused,   // program limit exceeded: an array value nested past six dimensions
	"42883": filterValueRefused,   // no such operator: uuid = integer
	"2201B": filterPatternRefused, // invalid regular expression
	"22025": filterPatternRefused, // invalid escape sequence: a LIKE pattern ending in its escape
	"2200C": filterPatternRefused, // invalid use of escape character: a SIMILAR TO pattern with a third escape separator
}

// filterRefusal turns Postgres's refusal of the caller's filter into the
// caller's error, and returns anything else as it came. clause is the filter
// as the list wrote it into statement ("WHERE (…)" or "AND (…)"), empty when
// the list had no filter; every list calls this where its query and its rows
// return their error, before handlePgError.
//
// Only with a filter: nothing else in a list's statement is the caller's
// text. The sort is re-parsed to "alias.column ASC|DESC", the fields come
// from the repository's own column list, and the cursor is built from the
// row it names. So a refusal of this kind on a statement with no filter is
// ours, and stays a fault.
//
// It returns an *domain.InvalidInputError on the filter, which every list
// handler answers with a 400. It does not wrap the driver's error: the
// message that goes out is the chain's, and Postgres's text must not be in
// it.
func filterRefusal(clause, statement string, err error) error {
	if clause == "" || err == nil {
		return err
	}

	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if !ok {
		return err
	}

	message, refused := filterRefusals[pgErr.Code]
	if !refused && syntaxErrorInFilter(clause, statement, pgErr) {
		message, refused = filterExpressionRefused, true
	}

	if !refused {
		return err
	}

	return &domain.InvalidInputError{Field: domain.FieldFilter, Message: message}
}

// syntaxErrorInFilter reports whether Postgres could not parse the statement
// at a place inside the caller's filter.
//
// A syntax error cannot be mapped by its code alone, as the others are: a
// statement this package built wrongly raises the same one, and that is
// ours. Postgres says where it stopped (Position, a character offset into
// the statement, from 1). When that place is inside the text the caller's
// filter became, the caller wrote the error; anywhere else, we did.
func syntaxErrorInFilter(clause, statement string, pgErr *pgconn.PgError) bool {
	if pgErr.Code != pgSyntaxError || pgErr.Position <= 0 {
		return false
	}

	// The parenthesised filter, without the WHERE or AND that joins it.
	open := strings.Index(clause, "(")
	if open < 0 {
		return false
	}

	written := clause[open:]

	before, _, ok := strings.Cut(statement, written)
	if !ok {
		return false
	}

	// Position counts characters, not bytes.
	first := utf8.RuneCountInString(before) + 1
	last := first + utf8.RuneCountInString(written) - 1
	position := int(pgErr.Position)

	return position >= first && position <= last
}
