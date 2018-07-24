package syncer

import (
	"errors"
	"fmt"
	"strings"

	re_type_v1alpha1 "github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/remoteenvironment/v1alpha1"
)

type reCRValidator struct{}

const (
	apiEntryType   = "API"
	eventEntryType = "Event"
)

// Validate validates RemoteEnvironment custom resource.
func (*reCRValidator) Validate(dto *re_type_v1alpha1.RemoteEnvironment) error {
	messages := []string{}

	for _, svc := range dto.Spec.Services {
		if len(svc.Entries) == 0 {
			messages = append(messages, fmt.Sprintf("Service with id %q is invalid. Entries list cannot be empty", svc.ID))
			continue
		}

		APIEntryCnt := 0
		EventEntryCnt := 0
		for _, entry := range svc.Entries {
			switch entry.Type {
			case apiEntryType:
				APIEntryCnt++

				if entry.GatewayUrl == "" {
					messages = append(messages, "GatewayUrl field is required for API type")
				}
				if entry.AccessLabel == "" {
					messages = append(messages, "AccessLabel field is required for API type")
				}
			case eventEntryType:
				EventEntryCnt++
			}
		}

		if APIEntryCnt > 1 {
			messages = append(messages, fmt.Sprintf("Service with id %q is invalid. Only one element with type API is allowed but found %d", svc.ID, APIEntryCnt))
		}

		if EventEntryCnt > 1 {
			messages = append(messages, fmt.Sprintf("Service with id %q is invalid. Only one element with type Event is allowed but found %d", svc.ID, EventEntryCnt))
		}
	}

	if len(messages) > 0 {
		return errors.New(strings.Join(messages, ", "))
	}

	return nil
}
