package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
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

	})

	t.Run("should return an error and default duration on invalid time unit", func(t *testing.T) {

	})

	t.Run("should return an error and default duration on invalid time value", func(t *testing.T) {

	})
}
