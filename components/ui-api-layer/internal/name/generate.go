package name

import (
	"strings"

	"github.com/moby/moby/pkg/namesgenerator"
)

func Generate() string {
	n := namesgenerator.GetRandomName(0)
	return strings.Replace(n, "_", "-", -1)
}
