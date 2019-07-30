package compass

import (
	"fmt"
)

type queryProvider struct{}

// TODO: createApplication does not support paging
// it will return only API, EventAPI and Document ids from page one
func (qp queryProvider) createApplication(input string) string {
	return fmt.Sprintf(`mutation {
	result: createApplication(in: %s) {
		id
		apis {
			data {
				id
			}
		}
		eventAPIs {
			data {
				id
			}
		}
		documents {
			data {
				id
			}
		}
	}
}`, input)
}

func (qp queryProvider) deleteApplication(id string) string {
	return fmt.Sprintf(`mutation {
	result: deleteApplication(id: "%s") {
		id
	}
}`, id)
}
