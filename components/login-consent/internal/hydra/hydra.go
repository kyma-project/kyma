package hydra

import (
	"context"
	"time"
)

func GetLoginRequest(ctx context.Context, ch chan<- string, challenge string) {
	select {
	case <-time.After(time.Second * 3):
		ch <- "Successful result."
	case <-ctx.Done():
		close(ch)
	}
}
