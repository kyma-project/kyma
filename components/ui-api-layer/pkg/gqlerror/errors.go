package gqlerror

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

type GQLError struct {
	kind      fmt.Stringer
	status    Status
	arguments map[string]string
	details   string
	message   string
}

func (e *GQLError) Status() Status {
	return e.status
}

type Status int

const (
	Unknown Status = iota
	Internal
	NotFound
	AlreadyExists
)

func (r Status) String() string {
	switch r {
	case NotFound:
		return "not found"
	case AlreadyExists:
		return "already exists"
	case Internal:
		return "internal error"
	default:
		return "unknown"
	}
}

type Option func(*GQLError)

func (e *GQLError) Error() string {
	return e.message
}

func New(err error, kind fmt.Stringer, opts ...Option) error {
	if err == nil {
		return nil
	}

	err = errors.Cause(err)

	switch {
	case apierrors.IsNotFound(err):
		return NewNotFound(kind, opts...)
	case apierrors.IsAlreadyExists(err):
		return NewAlreadyExists(kind, opts...)
	default:
		return NewInternal(opts...)
	}
}

func WithName(name string) Option {
	return func(gqlError *GQLError) {
		gqlError.arguments["name"] = name
	}
}

func WithEnvironment(environment string) Option {
	return func(gqlError *GQLError) {
		gqlError.arguments["environment"] = environment
	}
}

func WithCustomArgument(argument, value string) Option {
	return func(gqlError *GQLError) {
		gqlError.arguments[argument] = value
	}
}

func WithDetails(details string) Option {
	return func(gqlError *GQLError) {
		gqlError.details = details
	}
}

func NewInternal(opts ...Option) error {
	return buildError(nil, Internal, opts...)
}

func NewNotFound(kind fmt.Stringer, opts ...Option) error {
	return buildError(kind, NotFound, opts...)
}

func NewAlreadyExists(kind fmt.Stringer, opts ...Option) error {
	return buildError(kind, AlreadyExists, opts...)
}

func IsNotFound(err error) bool {
	return statusForError(err) == NotFound
}

func IsAlreadyExists(err error) bool {
	return statusForError(err) == AlreadyExists
}

func IsInternal(err error) bool {
	return statusForError(err) == Internal
}

func statusForError(err error) Status {
	type errorWithStatus interface {
		Status() Status
	}

	switch t := err.(type) {
	case errorWithStatus:
		return t.Status()
	}

	return Unknown
}

func buildError(kind fmt.Stringer, reason Status, opts ...Option) *GQLError {
	err := GQLError{kind: kind, status: reason, arguments: make(map[string]string, 0)}
	for _, opt := range opts {
		opt(&err)
	}

	err.message = buildMessage(&err)
	return &err
}

func buildMessage(err *GQLError) string {
	message := ""
	if err.kind != nil && !IsInternal(err) {
		message = fmt.Sprintf("%s ", err.kind)
	}

	message += fmt.Sprintf("%s", err.status)

	if len(err.arguments) > 0 {
		message += fmt.Sprintf(" [%s]", buildArguments(err.arguments))
	}

	if err.details != "" {
		message += fmt.Sprintf(": %s", err.details)
	}

	return message
}

func buildArguments(arguments map[string]string) string {
	keys := make([]string, len(arguments))

	for k := range arguments {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := ""
	for _, key := range keys {
		result = appendArgument(result, key, arguments[key])
	}

	return result
}

func appendArgument(arguments, name, value string) string {
	result := arguments
	if value != "" {
		if result != "" {
			result += ", "
		}

		result += fmt.Sprintf("%s: %q", name, value)
	}

	return result
}
