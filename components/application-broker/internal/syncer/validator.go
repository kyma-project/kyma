package syncer

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/application-broker/internal"

	appTypes "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
)

const (
	connectedAppLabelKey = "connected-app"
)

type appCRValidatorV2 struct{}

// Validate validates Application custom resource.
func (v *appCRValidatorV2) Validate(app *appTypes.Application) error {
	var messages []string

	for _, svc := range app.Spec.Services {
		if len(svc.Entries) == 0 {
			messages = append(messages, fmt.Sprintf("Service with id %q is invalid. Entries list cannot be empty", svc.ID))
			continue
		}

		apiEntryCnt := 0
		eventEntryCnt := 0
		for _, entry := range svc.Entries {
			switch entry.Type {
			case internal.APIEntryType:
				apiEntryCnt++

				if entry.GatewayUrl == "" {
					messages = append(messages, fmt.Sprintf("Service with id %q is invalid. GatewayUrl field is required for API type", svc.ID))
				}
				if entry.TargetUrl == "" {
					messages = append(messages, fmt.Sprintf("Service with id %q is invalid. TargetUrl field is required for API type", svc.ID))
				}
				if entry.Name == "" {
					messages = append(messages, fmt.Sprintf("Service with id %q is invalid. Name field is required for API type", svc.ID))
				}

			case internal.EventEntryType:
				eventEntryCnt++
			default:
				messages = append(messages, fmt.Sprintf("Service with id %q is invalid. Unknow entry type %q", svc.ID, entry.Type))
			}
		}

		if apiEntryCnt == 0 && eventEntryCnt == 0 {
			messages = append(messages, fmt.Sprintf("Service with id %q is invalid. Requires at least one API or Event entry", svc.ID))
		}
	}
	if !v.containsConnectedAppLabel(app.Spec.Labels) {
		messages = append(messages, fmt.Sprintf("Application %q is invalid. Labels field does not contains %s entry", app.Name, connectedAppLabelKey))
	}

	if app.Spec.CompassMetadata == nil || app.Spec.CompassMetadata.ApplicationID == "" {
		messages = append(messages, fmt.Sprintf("Application %q is invalid. CompassMetadata.ApplicationID field cannot be empty", app.Name))
	}

	if len(messages) > 0 {
		return errors.New(strings.Join(messages, ", "))
	}

	return nil
}

func (*appCRValidatorV2) containsConnectedAppLabel(labels map[string]string) bool {
	for key := range labels {
		if key == connectedAppLabelKey {
			return true
		}
	}
	return false
}

// Deprecated, remove in https://github.com/kyma-project/kyma/issues/7415
type appCRValidator struct{}

// Validate validates Application custom resource.
func (v *appCRValidator) Validate(dto *appTypes.Application) error {
	var messages []string

	for _, svc := range dto.Spec.Services {
		if len(svc.Entries) == 0 {
			messages = append(messages, fmt.Sprintf("Service with id %q is invalid. Entries list cannot be empty", svc.ID))
			continue
		}

		APIEntryCnt := 0
		EventEntryCnt := 0
		for _, entry := range svc.Entries {
			switch entry.Type {
			case internal.APIEntryType:
				APIEntryCnt++

				if entry.GatewayUrl == "" {
					messages = append(messages, "GatewayUrl field is required for API type")
				}
				if entry.AccessLabel == "" {
					messages = append(messages, "AccessLabel field is required for API type")
				}
			case internal.EventEntryType:
				EventEntryCnt++
			}
		}

		if APIEntryCnt > 1 {
			messages = append(messages, fmt.Sprintf("Service with id %q is invalid. Only one element with type API is allowed but found %d", svc.ID, APIEntryCnt))
		}

		if EventEntryCnt > 1 {
			messages = append(messages, fmt.Sprintf("Service with id %q is invalid. Only one element with type Event is allowed but found %d", svc.ID, EventEntryCnt))
		}

		if !v.containsConnectedAppLabel(svc.Labels) {
			messages = append(messages, fmt.Sprintf("Service with id %q is invalid. Labels field does not contains %s entry", svc.ID, connectedAppLabelKey))
		}
	}

	if len(messages) > 0 {
		return errors.New(strings.Join(messages, ", "))
	}

	return nil
}

func (*appCRValidator) containsConnectedAppLabel(labels map[string]string) bool {
	for key := range labels {
		if key == connectedAppLabelKey {
			return true
		}
	}
	return false
}
