package retrier

import (
	"fmt"
	"log"

	"github.com/pkg/errors"
)

const (
	UpdateRetries = 5
)

func Retry(fn func() error, retries int) error {
	var err error
	for i := 0; i <= retries; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		log.Println(err)
	}

	return errors.Wrapf(err, fmt.Sprintf("passed function failed after %d retries", retries))
}
