package cleaner

import "time"

// WithTimeProvider allows for passing custom time provider.
func (c *AzureCleaner) WithTimeProvider(nowProvider func() time.Time) *AzureCleaner {
	c.nowProvider = nowProvider
	return c
}
