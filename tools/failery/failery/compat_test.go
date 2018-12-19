package failery

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/tools/failery/failery/fixtures/mocks"
	"github.com/stretchr/testify/suite"
)

// CompatSuite covers compatbility with github.com/stretchr/testify/mock.
type CompatSuite struct {
	suite.Suite
}

// TestFailing asserts that methods like Mock.On accept variadic arguments
// that mirror those of the subject call.
func (s *CompatSuite) TestReturnError() {
	t := s.T()
	expectedErr := errors.New("Expected Error")
	m := mocks.NewRequesterVariadic(expectedErr)
	res, err := m.Sprintf("int: %d string: %s", 22, "twenty two")

	assert.Empty(t, res)
	assert.Equal(t, expectedErr, err)
}

func (s *CompatSuite) TestReturnErrorPointer() {
	t := s.T()
	expectedErr := errors.New("Expected Error")
	m := mocks.NewRequesterVariadic(expectedErr)
	res, err := m.Get("int: %d string: %s")

	assert.Empty(t, res)
	assert.Equal(t, &expectedErr, err)
}

func TestCompatSuite(t *testing.T) {
	mockcompatSuite := new(CompatSuite)
	suite.Run(t, mockcompatSuite)
}
