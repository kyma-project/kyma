package applications

import "github.com/stretchr/testify/require"

type Comparator interface {
	Compare(actualApp, expectedApp string) error
}

func NewComparator(assertions *require.Assertions) (Comparator, error) {
	return comparator{}, nil
}

type comparator struct {
	assertions *require.Assertions
}

func (c comparator) Compare(actualApp, expectedApp string) error {
	return nil
}
