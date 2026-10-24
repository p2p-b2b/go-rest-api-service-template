package domain

import (
	"errors"
	"fmt"
	"regexp"
	"slices"

	"github.com/slashdevops/qfv"
)

// ListExpressionCode is the one code a refused sort, filter or fields
// expression carries. The ValidationError's Field (sort, filter, fields) says
// which parameter it was.
const ListExpressionCode = "INVALID_EXPRESSION"

// listExpressions is one list request's three expression parameters, with
// the parsers that judge them and the allow-lists those parsers were built
// from. The allow-lists are here because the wording below is decided from
// them, not from the parser's error text.
type listExpressions struct {
	sortParser   *qfv.SortParser
	filterParser *qfv.FilterParser
	fieldsParser *qfv.FieldsParser

	sort, filter, fields string

	sortFields, filterFields, partialFields []string
}

// validateListExpressions parses the three expressions and records one
// ValidationError per parameter that does not parse. It is the only caller
// of the qfv parsers (TestListExpressionsAreParsedInOnePlace).
//
// # Our wording, not the library's
//
// Every list validator copied the qfv error into the response ("error on
// field 'x': field not allowed for sorting"), under four codes for the same
// failure (INVALID_SORT_FIELD, INVALID_FILTER_FIELD, INVALID_FIELD,
// INVALID_FIELDS_FIELD). That published a dependency's prose as the API
// contract: a qfv release could reword every 400 of every list endpoint.
// The message is now ours: it names the field when the expression refers to
// a field this list does not accept, and otherwise says what the expression
// must look like.
//
// The field is named only when it is a plain identifier missing from the
// allow-list we hold, so the decision does not depend on the library's
// message text, and nothing but a short identifier is echoed back.
//
// # An expression over its length is not parsed
//
// A list's sort, filter and fields arrive in the query string, which the
// body limit does not bound: a request line may be as long as the header
// limit allows, 1 MiB by default. The length check and the parse were two
// independent steps, and the first was each input's to remember: six list
// inputs recorded TOO_LONG and handed the expression to the parser anyway,
// and four had no length check at all. The filter parser is recursive, and
// until qfv v1.0.3 the input chose how deep it recursed: a deeply nested
// filter overflowed the stack, which is a fatal error of the runtime that
// Recovery does not catch.
//
// The parser bounds its own depth now. The length bound is still enforced
// here, in front of it, where the parsers are called: it is the cheaper
// refusal, it covers the sort and the fields, and it does not depend on
// what a library does with a megabyte of input. An expression over its
// length is refused as TOO_LONG, once, and no parser sees it. No input
// checks the length for itself any more: there is one place to get it
// right.
func validateListExpressions(errs *ValidationErrors, in listExpressions) {
	in.sort = boundedExpression(errs, in.sort, FieldSort, ValidateSortExpression)
	in.filter = boundedExpression(errs, in.filter, FieldFilter, ValidateFilterExpression)
	in.fields = boundedExpression(errs, in.fields, FieldFields, ValidateFieldsExpression)

	if in.sort != "" {
		if _, err := in.sortParser.Parse(in.sort); err != nil {
			errs.Add(listExpressionError(FieldSort, err, in.sortFields,
				"'%s' is not a sortable field",
				"must be a comma-separated list of 'field asc' or 'field desc'"))
		}
	}

	if in.filter != "" {
		if _, err := in.filterParser.Parse(in.filter); err != nil {
			errs.Add(listExpressionError(FieldFilter, err, in.filterFields,
				"'%s' is not a filterable field",
				"is not a valid filter expression"))
		}
	}

	if in.fields != "" {
		if _, err := in.fieldsParser.Parse(in.fields); err != nil {
			errs.Add(listExpressionError(FieldFields, err, in.partialFields,
				"'%s' is not a selectable field",
				"must be a comma-separated list of field names"))
		}
	}
}

// boundedExpression returns the expression a parser may be given: the
// expression itself, or nothing when it is over its length, which is
// recorded.
func boundedExpression(errs *ValidationErrors, expression, field string, check func(expression, field string) error) string {
	if err := check(expression, field); err != nil {
		errs.Add(err)

		return ""
	}

	return expression
}

// listIdentifier is what may be echoed back as a field name: short, and
// nothing a client could use to smuggle markup or control characters into a
// response or a log line.
var listIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.]{0,63}$`)

func listExpressionError(field string, err error, allowed []string, unknownFormat, invalid string) *ValidationError {
	for _, name := range qfvFieldNames(err) {
		if listIdentifier.MatchString(name) && !slices.Contains(allowed, name) {
			return &ValidationError{Field: field, Message: fmt.Sprintf(unknownFormat, name), Code: ListExpressionCode}
		}
	}

	return &ValidationError{Field: field, Message: invalid, Code: ListExpressionCode}
}

// qfvFieldNames returns the Field of every qfv error in err, in order. The
// filter parser joins several errors; the sort and fields parsers return one.
func qfvFieldNames(err error) []string {
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		var names []string
		for _, e := range joined.Unwrap() {
			names = append(names, qfvFieldNames(e)...)
		}

		return names
	}

	if e, ok := errors.AsType[*qfv.QFVSortError](err); ok {
		return []string{e.Field}
	}

	if e, ok := errors.AsType[*qfv.QFVFilterError](err); ok {
		return []string{e.Field}
	}

	if e, ok := errors.AsType[*qfv.QFVFieldsError](err); ok {
		return []string{e.Field}
	}

	return nil
}
