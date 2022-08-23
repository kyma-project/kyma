package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseInvalidSections(t *testing.T) {
	section := "invalid"
	_, err := ParseCustomSection(section)
	require.Error(t, err)
}

func TestEmptySection(t *testing.T) {
	section := ""
	_, err := ParseCustomSection(section)
	require.NoError(t, err)
}

func TestParseInvalidMultiLineSections(t *testing.T) {
	section := "key value\ninvalid"
	_, err := ParseCustomSection(section)
	require.Error(t, err)
}

func TestParseSingleLine(t *testing.T) {
	section := "key value"
	result, err := ParseCustomSection(section)
	require.NoError(t, err)
	require.True(t, result.ContainsKey("key"))
	require.NotNil(t, result.GetByKey("key"))
	require.Equal(t, result.GetByKey("key").Value, "value")
}

func TestParseUpperCaseLine(t *testing.T) {
	section := "Key value"
	result, err := ParseCustomSection(section)
	require.NoError(t, err)
	require.False(t, result.ContainsKey("Key"))
	require.NotNil(t, result.GetByKey("key"))
	require.Equal(t, result.GetByKey("key").Value, "value")
}

func TestParseWithSpaces(t *testing.T) {
	section := "key value1 value2"
	result, err := ParseCustomSection(section)
	require.NoError(t, err)
	require.True(t, result.ContainsKey("key"))
	require.NotNil(t, result.GetByKey("key"))
	require.Equal(t, result.GetByKey("key").Value, "value1 value2")
}

func TestParseUntrimmedLine(t *testing.T) {
	section := "   key  value   "
	result, err := ParseCustomSection(section)
	require.NoError(t, err)
	require.True(t, result.ContainsKey("key"))
	require.NotNil(t, result.GetByKey("key"))
	require.Equal(t, result.GetByKey("key").Value, "value")
}

func TestParseMultiLine(t *testing.T) {
	section := "key1 value1\nkey2 value2\n\nkey3 value3"
	result, err := ParseCustomSection(section)
	require.NoError(t, err)
	require.True(t, result.ContainsKey("key1"))
	require.NotNil(t, result.GetByKey("key1"))
	require.Equal(t, result.GetByKey("key1").Value, "value1")
	require.True(t, result.ContainsKey("key2"))
	require.NotNil(t, result.GetByKey("key2"))
	require.Equal(t, result.GetByKey("key2").Value, "value2")
	require.True(t, result.ContainsKey("key3"))
	require.NotNil(t, result.GetByKey("key3"))
	require.Equal(t, result.GetByKey("key3").Value, "value3")
}

func TestParseComment(t *testing.T) {
	section := "#comment"
	_, err := ParseCustomSection(section)
	require.NoError(t, err)
}

func TestParseMultiLineWithComment(t *testing.T) {
	section := "key1 value1\n#comment\nkey2 value2\n\nkey3 value3"
	result, err := ParseCustomSection(section)
	require.NoError(t, err)
	require.True(t, result.ContainsKey("key1"))
	require.NotNil(t, result.GetByKey("key1"))
	require.Equal(t, result.GetByKey("key1").Value, "value1")
	require.True(t, result.ContainsKey("key2"))
	require.NotNil(t, result.GetByKey("key2"))
	require.Equal(t, result.GetByKey("key2").Value, "value2")
	require.True(t, result.ContainsKey("key3"))
	require.NotNil(t, result.GetByKey("key3"))
	require.Equal(t, result.GetByKey("key3").Value, "value3")
}

func TestGrepFilterSection(t *testing.T) {
	section := "    Name   grep\n    Match   *\n    Regex   $kubernetes['labels']['app'] my-deployment"
	_, err := ParseCustomSection(section)
	require.NoError(t, err)
}

func TestRewriteTagFilterSection(t *testing.T) {
	section := "        Name   rewrite_tag\n        Match  kube.*\n        Rule   $log \"^.*$\" log_rewritten-1 true\n        Emitter_Name  log_emitter-1\n        Emitter_Storage.type filesystem"
	_, err := ParseCustomSection(section)
	require.NoError(t, err)
}
