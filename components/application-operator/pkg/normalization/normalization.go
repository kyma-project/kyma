package normalization

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var nonAlphaNumeric = regexp.MustCompile("[^A-Za-z0-9]+")

// NormalizeServiceNameWithId creates the OSB Service Name for given Application Service.
// The OSB Service Name is used in the Service Catalog as the clusterServiceClassExternalName, so it need to be normalized.
//
// Normalization rules:
// - MUST only contain lowercase characters, numbers and hyphens (no spaces).
// - MUST be unique across all service objects returned in this response. MUST be a non-empty string.
func NormalizeServiceNameWithId(displayName, id string) string {

	normalizedName := NormalizeName(displayName)

	// add suffix
	// generate 5 characters suffix from the id
	sha := sha1.New()
	sha.Write([]byte(id))
	suffix := hex.EncodeToString(sha.Sum(nil))[:5]
	normalizedName = fmt.Sprintf("%s-%s", normalizedName, suffix)
	// remove dash prefix if exists
	//  - can happen, if the name was empty before adding suffix empty or had dash prefix
	normalizedName = strings.TrimPrefix(normalizedName, "-")

	return normalizedName
}

func NormalizeName(displayName string) string {

	// remove all characters, which is not alpha numeric
	normalizedName := nonAlphaNumeric.ReplaceAllString(displayName, "-")
	// to lower
	normalizedName = strings.Map(unicode.ToLower, normalizedName)
	// trim dashes if exists
	normalizedName = strings.TrimSuffix(normalizedName, "-")
	if len(normalizedName) > 57 {
		normalizedName = normalizedName[:57]
	}

	return strings.TrimPrefix(normalizedName, "-")
}
