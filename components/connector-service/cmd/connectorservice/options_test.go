package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOptions_ParseEnv(t *testing.T) {
	t.Run("should parse environmental variables", func(t *testing.T) {
		//given
		env := environment{
			country:            "country",
			organization:       "organization",
			organizationalUnit: "organizationalunit",
			locality:           "locality",
			province:           "province",
		}

		os.Setenv("COUNTRY", env.country)
		os.Setenv("ORGANIZATION", env.organization)
		os.Setenv("ORGANIZATIONALUNIT", env.organizationalUnit)
		os.Setenv("LOCALITY", env.locality)
		os.Setenv("PROVINCE", env.province)

		//when
		res := parseEnv()

		//then
		assert.EqualValues(t, env, *res)
	})
}

func TestOptions_ParseDuration(t *testing.T) {
	t.Run("should parse proper duration string", func(t *testing.T) {
		//given
		durationString := "30h"

		//when
		res, err := parseDuration(durationString)

		//then
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(30)*time.Hour, res)
	})

	t.Run("should return an error and default duration on invalid time unit", func(t *testing.T) {
		//given
		invalidDurationString := "4u"

		//when
		res, err := parseDuration(invalidDurationString)

		//then
		assert.Equal(t, defaultCertificateValidityTime, res)
		assert.Error(t, err)
		assert.Equal(t, "unrecognized time unit provided: u", err.Error())
	})

	t.Run("should return an error and default duration on invalid time value", func(t *testing.T) {
		//given
		invalidDurationString := "abcdh"

		//when
		res, err := parseDuration(invalidDurationString)

		//then
		assert.Equal(t, defaultCertificateValidityTime, res)
		assert.Error(t, err)
		assert.Equal(t, "strconv.Atoi: parsing \"abcd\": invalid syntax", err.Error())
	})
}

func TestParseNamespacedName(t *testing.T) {
	testCases := []struct {
		value     string
		namespace string
		name      string
	}{
		{
			value:     "kyma-integration/ca-secret",
			namespace: "kyma-integration",
			name:      "ca-secret",
		},
		{
			value:     "ca-secret",
			namespace: defaultNamespace,
			name:      "ca-secret",
		},
		{
			value:     "/ca-secret",
			namespace: defaultNamespace,
			name:      "ca-secret",
		},
		{
			value:     "ca-secret/",
			namespace: defaultNamespace,
			name:      "ca-secret",
		},
		{
			value:     "/",
			namespace: defaultNamespace,
			name:      "",
		},
	}

	for _, test := range testCases {
		t.Run(fmt.Sprintf("should parse \"%s\" namespaced name", test.value), func(t *testing.T) {
			namespaceName := parseNamespacedName(test.value)
			assert.Equal(t, test.namespace, namespaceName.Namespace)
			assert.Equal(t, test.name, namespaceName.Name)
		})
	}
}
