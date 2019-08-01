package v1_test

import (
	asyncapierror "github.com/asyncapi/converter-go/pkg/error"
	v1 "github.com/kyma-project/kyma/components/cms-services/pkg/endpoint/asyncapi/v1"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"golang.org/x/net/context"

	"io"
	"io/ioutil"
	"strings"
	"testing"
)

var testConvert = v1.Convert(func(reader io.Reader, writer io.Writer) error {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	_, err = writer.Write(bytes)
	if err != nil {
		return err
	}
	return nil
})

var errTest = errors.New("test error")

type failingReader struct{}

func (failingReader) Read(p []byte) (n int, err error) {
	return 0, errTest
}

type noModificationReader struct{}

func (noModificationReader) Read(p []byte) (n int, err error) {
	return 0, asyncapierror.NewDocumentVersionUpToDate(nil)
}

func TestConvert_Mutate(t *testing.T) {
	g := NewWithT(t)
	reader := strings.NewReader("test me plz")
	bytes, modified, err := testConvert.Mutate(context.TODO(), reader, "")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(bytes).To(Equal([]byte("test me plz")))
	g.Expect(modified).To(BeTrue())
}

func TestConvert_Mutate_reader_no_modification(t *testing.T) {
	g := NewWithT(t)
	_, modified, err := testConvert.Mutate(context.TODO(), noModificationReader{}, "")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(modified).To(BeFalse())
}

func TestConvert_Mutate_reader_err(t *testing.T) {
	g := NewWithT(t)
	_, modified, err := testConvert.Mutate(context.TODO(), failingReader{}, "")
	g.Expect(err).To(HaveOccurred())
	g.Expect(modified).To(BeFalse())
}
