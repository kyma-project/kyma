package logger_test

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/components/application-connectivity-validator/pkg/logger"
)

func TestMapLevel(t *testing.T) {
	in := ""

	out, err := logger.MapLevel(in)

	fmt.Println(out)
	fmt.Println(err)
}
