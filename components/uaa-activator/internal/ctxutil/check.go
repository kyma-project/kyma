package ctxutil

import "context"

// ShouldExit returns true if context is canceled
func ShouldExit(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
