// Package time provide features which supplements standard time package.
package time

import gTime "time"

// NowProvider is a provider for current time
type NowProvider func() gTime.Time

// Now returns current time.
// If NowProvider is not initialised, then current time is returned,
// otherwise NowProvider is called and result returned.
func (m NowProvider) Now() gTime.Time {
	if m == nil {
		return gTime.Now()
	}
	return m()
}
