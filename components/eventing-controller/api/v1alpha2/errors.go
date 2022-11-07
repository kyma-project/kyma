package v1alpha2

import (
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

//nolint:gochecknoglobals // these are required for testing
var (
	SourcePath = field.NewPath("spec").Child("source")
	TypesPath  = field.NewPath("spec").Child("types")
	ConfigPath = field.NewPath("spec").Child("config")
	SinkPath   = field.NewPath("spec").Child("sink")
	NSPath     = field.NewPath("metadata").Child("namespace")

	EmptyErrDetail          = "must not be empty"
	DuplicateTypesErrDetail = "must not have duplicate types"
	LengthErrDetail         = "must not be of length zero"
	MinSegmentErrDetail     = fmt.Sprintf("must have minimum %s segments", strconv.Itoa(minEventTypeSegments))
	InvalidPrefixErrDetail  = fmt.Sprintf("must not have %s as type prefix", InvalidPrefix)
	StringIntErrDetail      = fmt.Sprintf("%s must be a stringified int value", MaxInFlightMessages)

	MissingSchemeErrDetail = "must have URL scheme 'http' or 'https'"
	SuffixMissingErrDetail = fmt.Sprintf("must have valid sink URL suffix %s", ClusterLocalURLSuffix)
	SubDomainsErrDetail    = fmt.Sprintf("must have sink URL with %d sub-domains: ", subdomainSegments)
	NSMismatchErrDetail    = "must have the same namespace as the subscriber: "
)

func MakeInvalidFieldError(path *field.Path, subName, detail string) *field.Error {
	return field.Invalid(path, subName, detail)
}
