package exit

import (
	"log"

	"github.com/pkg/errors"
)

func OnError(err error, context string, args ...interface{}) {
	if err == nil {
		return
	}

	log.Fatal(errors.Wrapf(err, context, args...))
}
