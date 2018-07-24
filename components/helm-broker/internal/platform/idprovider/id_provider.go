package idprovider

//noinspection SpellCheckingInspection
import (
	crand "crypto/rand"
	"time"

	"github.com/oklog/ulid"
	"github.com/pkg/errors"
)

// New returns function which generates ULID ids.
// Reader from crypto/rand is used as a entropy source. It is safe for concurrent use.
func New() func() (string, error) {
	return func() (string, error) {
		ulidGen, err := ulid.New(ulid.Timestamp(time.Now()), crand.Reader)
		if err != nil {
			// not covered directly by tests due to quite difficult trigger scenario.
			return "", errors.Wrap(err, "while generating ID")
		}
		return ulidGen.String(), nil
	}
}
