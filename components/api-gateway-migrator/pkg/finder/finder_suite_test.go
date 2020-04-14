package finder_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFinder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Finder Suite")
}
