package fluentbit

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseInvalidSections(t *testing.T) {
	section := "invalid"
	_, err := ParseSection(section)
	require.Error(t, err)
}

func TestEmptySection(t *testing.T) {
	section := ""
	result, err := ParseSection(section)
	require.NoError(t, err)
	require.Equal(t, map[string]string{}, result)
}

func TestParseInvalidMultiLineSections(t *testing.T) {
	section := "key value\ninvalid"
	_, err := ParseSection(section)
	require.Error(t, err)
}

func TestParseSingleLine(t *testing.T) {
	section := "key value"
	result, err := ParseSection(section)
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "value"}, result)
}

func TestParseUpperCaseLine(t *testing.T) {
	section := "Key value"
	result, err := ParseSection(section)
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "value"}, result)
}

func TestParseWithSpaces(t *testing.T) {
	section := "key value1 value2"
	result, err := ParseSection(section)
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "value1 value2"}, result)
}

func TestParseUntrimmedLine(t *testing.T) {
	section := "   key  value   "
	result, err := ParseSection(section)
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "value"}, result)
}

func TestParseMultiLine(t *testing.T) {
	section := "key1 value1\nkey2 value2\n\nkey3 value3"
	result, err := ParseSection(section)
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key1": "value1", "key2": "value2", "key3": "value3"}, result)
}

func TestParseComment(t *testing.T) {
	section := "#comment"
	result, err := ParseSection(section)
	require.NoError(t, err)
	require.Equal(t, map[string]string{}, result)
}

func TestParseMultiLineWithComment(t *testing.T) {
	section := "key1 value1\n#comment\nkey2 value2\n\nkey3 value3"
	result, err := ParseSection(section)
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key1": "value1", "key2": "value2", "key3": "value3"}, result)
}

func TestGrepFilterSection(t *testing.T) {
	section := "    Name   grep\n    Match   *\n    Regex   $kubernetes['labels']['app'] my-deployment"
	result, err := ParseSection(section)
	require.NoError(t, err)
	require.Equal(t, map[string]string{
		"name":  "grep",
		"match": "*",
		"regex": "$kubernetes['labels']['app'] my-deployment",
	}, result)

	name, err := getSectionName(result)
	require.NoError(t, err)
	require.Equal(t, "grep", name)
}

func TestRewriteTagFilterSection(t *testing.T) {
	section := "        Name   rewrite_tag\n        Match  kube.*\n        Rule   $log \"^.*$\" log_rewritten-1 true\n        Emitter_Name  log_emitter-1\n        Emitter_Storage.type filesystem"
	result, err := ParseSection(section)
	require.NoError(t, err)
	require.Equal(t, map[string]string{
		"name":                 "rewrite_tag",
		"match":                "kube.*",
		"rule":                 "$log \"^.*$\" log_rewritten-1 true",
		"emitter_name":         "log_emitter-1",
		"emitter_storage.type": "filesystem",
	}, result)

	name, err := getSectionName(result)
	require.NoError(t, err)
	require.Equal(t, "rewrite_tag", name)
}

func getSectionName(section map[string]string) (string, error) {
	if name, hasKey := section["name"]; hasKey {
		return name, nil
	}
	return "", fmt.Errorf("configuration section does not have name attribute")
}
