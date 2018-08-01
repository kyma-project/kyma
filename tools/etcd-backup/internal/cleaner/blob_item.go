package cleaner

import "time"

type simpleBlobItem struct {
	Name         string
	LastModified time.Time
}

// byLastModifiedDesc implements sort.Interface and allows you to sort the slice of simpleBlobItem
// by LastModified property in a descending order.
type byLastModifiedDesc []simpleBlobItem

func (b byLastModifiedDesc) Len() int           { return len(b) }
func (b byLastModifiedDesc) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byLastModifiedDesc) Less(i, j int) bool { return b[i].LastModified.After(b[j].LastModified) }
