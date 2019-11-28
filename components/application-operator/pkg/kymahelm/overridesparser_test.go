package kymahelm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testTemplate = `global:
  domainName: {{ .DomainName }}
  testValue: {{ .TestValue }}`
)

func TestParseOverrides(t *testing.T) {

	t.Run("should parse overrides", func(t *testing.T) {
		// given
		expectedOverrides := `global:
  domainName: my.domain
  testValue: 100`

		data := struct {
			DomainName string
			TestValue  int
		}{
			DomainName: "my.domain",
			TestValue:  100,
		}

		// when
		overrides, err := ParseOverrides(data, testTemplate)

		// then
		assert.NoError(t, err)
		assert.Equal(t, expectedOverrides, overrides)
	})

	t.Run("should return error if invalid data provided", func(t *testing.T) {
		// given
		data := struct {
			WrongData string
		}{
			WrongData: "wrongData",
		}

		// when
		overrides, err := ParseOverrides(data, testTemplate)

		// then
		assert.Error(t, err)
		assert.Equal(t, "", overrides)
	})
}
