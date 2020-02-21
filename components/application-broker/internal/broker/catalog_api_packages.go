package broker

var appYAML = `name: mocked
description: "EC description"
tags:
  - occ
  - promotions
labels:
  connected-app: mocked
providerDisplayName: SAP
longDescription: These services are used to get search profiles using sort, facet, and boost settings.
displayName: SAP Commerce Cloud
compassMetadata:
  applicationId: "28d0bfa1-3b2d-45c4-b26b-24406f517ece" # unique
services:
  - description: These services are used to get search profiles using sort, facet,
      and boost settings.
    displayName: SAP Commerce Cloud - Adaptive Search Webservices
    entries:
      - gatewayUrl: http://mocked-cd96a180-f147-4167-ae4d-4a6c3f3742c2.kyma-integration.svc.cluster.local
        targetUrl: https://commerce-mocks.api-packages-mock.cluster.extend.cx.cloud.sap/adaptivesearchwebservices
        type: API
        name: ADAPTIVE_SEARCH_WEBSERVICES
    id: cd96a180-f147-4167-ae4d-4a6c3f3742c2
    name: sap-commerce-cloud-adaptive-search-webservices-1c94e
  - description: These enhanced services manage CMS-related items specifically designed
      for SmartEdit.
    displayName: SAP Commerce Cloud - CMS SmartEdit Webservices
    entries:
      - gatewayUrl: http://mocked-fb5d2121-aa30-49d2-a345-a8b59eee31ec.kyma-integration.svc.cluster.local
        targetUrl: https://commerce-mocks.api-packages-mock.cluster.extend.cx.cloud.sap/cmssmarteditwebservices
        type: API
        name: SmartEdit
    id: fb5d2121-aa30-49d2-a345-a8b59eee31ec
    name: sap-commerce-cloud-cms-smartedit-webservices-b2c86
  - description: These client-independent services manage CMS-related items.
    displayName: SAP Commerce Cloud - CMS Webservices
    entries:
      - gatewayUrl: http://mocked-c28f5d9e-daa0-4f15-9f45-f54b69e81031.kyma-integration.svc.cluster.local
        targetUrl: https://commerce-mocks.api-packages-mock.cluster.extend.cx.cloud.sap/cmswebservices
        type: API
        name: CMS
    id: c28f5d9e-daa0-4f15-9f45-f54b69e81031
    name: sap-commerce-cloud-commerce-webservices-12334
  - description: These services manage all of the common commerce functionality,
      and also include customizations from installed AddOns. The implementing extension
      is called ycommercewebservices.
    displayName: SAP Commerce Cloud - Commerce Webservices
    entries:
      - gatewayUrl: http://mocked-df24ca29-4caf-46de-a039-8f72e82108ad.kyma-integration.svc.cluster.local
        targetUrl: https://commerce-mocks.api-packages-mock.cluster.extend.cx.cloud.sap/rest/v2
        type: API
        name: Commerce
    id: df24ca29-4caf-46de-a039-8f72e82108ad
    name: sap-commerce-cloud-commerce-webservices-eb044
  - description: These services verify if a specific user or user group can access
      a type or an attribute, and if the user or user group has global or catalog-specific
      permissions.
    displayName: SAP Commerce Cloud - Permission Webservices
    entries:
      - gatewayUrl: http://mocked-b0ebcc7e-40c8-4124-badc-705a7d6a4abd.kyma-integration.svc.cluster.local
        targetUrl: https://commerce-mocks.api-packages-mock.cluster.extend.cx.cloud.sap/permissionswebservices
        type: API
        name: Permission
    id: b0ebcc7e-40c8-4124-badc-705a7d6a4abd
    name: sap-commerce-cloud-permission-webservices-68f0e
  - description: These services create and manage customizations targeted at specific
      users and can integrate personalization with other systems.
    displayName: SAP Commerce Cloud - Personalization Webservices
    entries:
      - gatewayUrl: http://mocked-573c6d1b-a6db-449a-a012-89cf7fd6b04a.kyma-integration.svc.cluster.local
        targetUrl: https://commerce-mocks.api-packages-mock.cluster.extend.cx.cloud.sap/personalizationwebservices
        type: API
        name: Personalization
    id: 573c6d1b-a6db-449a-a012-89cf7fd6b04a
    name: sap-commerce-cloud-personalization-webservices-147c8
  - description: These services create a preview ticket that contains session settings
      such as the catalog version, user data, or language data.
    displayName: SAP Commerce Cloud - Preview Webservices
    entries:
      - gatewayUrl: http://mocked-aabcfde1-7614-4614-b6b8-ac7f9bfcd11a.kyma-integration.svc.cluster.local
        targetUrl: https://commerce-mocks.api-packages-mock.cluster.extend.cx.cloud.sap/previewwebservices
        type: API
        name: Preview
    id: aabcfde1-7614-4614-b6b8-ac7f9bfcd11a
    name: sap-commerce-cloud-preview-webservices-e5b86
  - description: These services provide the ability for agents to assist customers
      with their order, and look up the order by name, email, cart, or order number.
    displayName: SAP Commerce Cloud - Assisted Service Webservices
    entries:
      - gatewayUrl: http://mocked-d53d0862-a1a2-4969-9643-39ed932faac0.kyma-integration.svc.cluster.local
        targetUrl: https://commerce-mocks.api-packages-mock.cluster.extend.cx.cloud.sap/assistedservicewebservices
        type: API
        name: Assisted
    id: d53d0862-a1a2-4969-9643-39ed932faac0
    name: sap-commerce-cloud-assisted-service-webservices-169d5
  - description: Set of events emitted typically by SAP Commerce Cloud
    displayName: SAP Commerce Cloud - Events
    entries:
      - type: Events
    id: 13ce0471-ab5b-4925-8446-55688e4700ee
    name: sap-commerce-cloud-assisted-service-webservices-12367
`

type CompassMetadata struct {
	ApplicationID string `json:"applicationId"`
}

// Application represents Application as defined by OSB API.
type Application struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Services    []Service `json:"services"`

	// New fields
	CompassMetadata     CompassMetadata   `json:"compassMetadata"`
	DisplayName         string            `json:"displayName"`
	ProviderDisplayName string            `json:"providerDisplayName"`
	LongDescription     string            `json:"longDescription"`
	Labels              map[string]string `json:"labels"`
}

// Service represents service defined in the application which is mapped to service class in the service catalog.
type Service struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`

	Entries []Entry `json:"entries"`
}

// Entry is a generic type for all type of entries.
type Entry struct {
	Type     string `json:"type"`
	APIEntry `json:",inline"`
}

// Entries represents API of the application.
type APIEntry struct {
	GatewayURL string
	TargetURL  string
	Name       string
}
