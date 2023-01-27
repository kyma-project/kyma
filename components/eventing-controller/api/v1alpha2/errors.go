package v1alpha2

import (
	"fmt"
	"strconv"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"

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
	InvalidURIErrDetail     = "must be valid as per RFC 3986"
	DuplicateTypesErrDetail = "must not have duplicate types"
	LengthErrDetail         = "must not be of length zero"
	MinSegmentErrDetail     = fmt.Sprintf("must have minimum %s segments", strconv.Itoa(minEventTypeSegments))
	InvalidPrefixErrDetail  = fmt.Sprintf("must not have %s as type prefix", InvalidPrefix)
	StringIntErrDetail      = fmt.Sprintf("%s must be a stringified int value", MaxInFlightMessages)

	InvalidQosErrDetail = fmt.Sprintf("must be a valid QoS value %s or %s",
		types.QosAtLeastOnce, types.QosAtMostOnce)
	InvalidAuthTypeErrDetail  = fmt.Sprintf("must be a valid Auth Type value %s", types.AuthTypeClientCredentials)
	InvalidGrantTypeErrDetail = fmt.Sprintf("must be a valid Grant Type value %s", types.GrantTypeClientCredentials)

	MissingSchemeErrDetail = "must have URL scheme 'http' or 'https'"
	SuffixMissingErrDetail = fmt.Sprintf("must have valid sink URL suffix %s", ClusterLocalURLSuffix)
	SubDomainsErrDetail    = fmt.Sprintf("must have sink URL with %d sub-domains: ", subdomainSegments)
	NSMismatchErrDetail    = "must have the same namespace as the subscriber: "
)

func MakeInvalidFieldError(path *field.Path, subName, detail string) *field.Error {
	return field.Invalid(path, subName, detail)
}
