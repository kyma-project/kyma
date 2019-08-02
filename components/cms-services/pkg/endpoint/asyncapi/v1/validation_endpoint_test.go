package v1_test

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	parser "github.com/asyncapi/parser/pkg"

	v1 "github.com/kyma-project/kyma/components/cms-services/pkg/endpoint/asyncapi/v1"
	"github.com/onsi/gomega"
)

func TestValidator_Validate(t *testing.T) {
	for testName, testCase := range map[string]struct {
		fail     bool
		filePath string
	}{
		"valid - yaml": {
			fail:     false,
			filePath: "./testdata/valid.yaml",
		},
		"invalid - yaml": {
			fail:     true,
			filePath: "./testdata/invalid.yaml",
		},
		"valid - json": {
			fail:     false,
			filePath: "./testdata/valid.json",
		},
		"invalid - json": {
			fail:     true,
			filePath: "./testdata/invalid.json",
		},
		"generic error - no parsing errors": {
			fail:     true,
			filePath: "",
		},
	} {
		t.Run(testName, func(t *testing.T) {
			// Given
			g := gomega.NewWithT(t)
			validator := v1.Validate(parser.Parse)
			var reader io.Reader
			if testCase.filePath != "" {
				file, err := os.Open(testCase.filePath)
				g.Expect(err).ToNot(gomega.HaveOccurred())
				defer file.Close()

				reader = file
			} else {
				reader = strings.NewReader("")
			}

			// When
			err := validator.Validate(context.TODO(), reader, "")

			// Then
			if testCase.fail {
				g.Expect(err).To(gomega.HaveOccurred())
			} else {
				g.Expect(err).ToNot(gomega.HaveOccurred())
			}
		})
	}
}
