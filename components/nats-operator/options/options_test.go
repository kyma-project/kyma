package options

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	// given
	const wantInterval = time.Second * 3
	os.Args = []string{"doctor", fmt.Sprintf("--%s=%v", argNameInterval, wantInterval)}

	// when
	opts := New()
	gotInterval := opts.Parse().Interval

	// expect
	assert.NotNil(t, opts)
	assert.Equal(t, wantInterval, gotInterval)
}
