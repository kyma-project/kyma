package v1alpha2

import (
	"fmt"
	"strings"

	kymaApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	k8sApiExtensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Crd(domainName string) *k8sApiExtensions.CustomResourceDefinition {

	kind := kymaApi.KindName
	listKind := kymaApi.ListKindName
	singular := strings.ToLower(kymaApi.KindName)
	plural := singular + "s"
	group := kymaApi.Group
	version := kymaApi.Version

	return &k8sApiExtensions.CustomResourceDefinition{
		ObjectMeta: k8sMeta.ObjectMeta{
			Name: fmt.Sprintf("%s.%s", plural, group),
		},
		Spec: k8sApiExtensions.CustomResourceDefinitionSpec{
			Group:   group,
			Version: version,
			Scope:   "Namespaced",
			Names: k8sApiExtensions.CustomResourceDefinitionNames{
				Singular: singular,
				Plural:   plural,
				Kind:     kind,
				ListKind: listKind,
			},
			Validation: &k8sApiExtensions.CustomResourceValidation{
				OpenAPIV3Schema: &k8sApiExtensions.JSONSchemaProps{
					Properties: map[string]k8sApiExtensions.JSONSchemaProps{
						"spec": {
							Required: []string{"service", "hostname"},
							Properties: map[string]k8sApiExtensions.JSONSchemaProps{
								"service": {
									Type:     "object",
									Required: []string{"name", "port"},
									Properties: map[string]k8sApiExtensions.JSONSchemaProps{
										"name": {
											Type: "string",
										},
										"port": {
											Type: "integer",
										},
									},
								},
								"hostname": {
									Type:      "string",
									Pattern:   hostnamePattern(domainName),
									MinLength: itoi64(3),
									MaxLength: itoi64(256),
								},
								"disableIstioAuthPolicyMTLS": {
									Type: "boolean",
								},
								"authenticationEnabled": {
									Type: "boolean",
								},
								"authentication": {
									Type: "array",
									Items: &k8sApiExtensions.JSONSchemaPropsOrArray{
										Schema: &k8sApiExtensions.JSONSchemaProps{
											Type:     "object",
											Required: []string{"type"},
											OneOf: []k8sApiExtensions.JSONSchemaProps{
												{Required: []string{"jwt"}},
											},
											Properties: map[string]k8sApiExtensions.JSONSchemaProps{
												"type": {
													Type: "string",
												},
												"jwt": {
													Type:     "object",
													Required: []string{"issuer", "jwksUri"},
													Properties: map[string]k8sApiExtensions.JSONSchemaProps{
														"issuer": {
															Type:    "string",
															Pattern: urlPattern,
														},
														"jwksUri": {
															Type:    "string",
															Pattern: urlPattern,
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

const (
	hostnamePatternFormat = `^([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]{0,62}[A-Za-z0-9])(\.%s)?$`
	urlPattern            = `^(?:https?:\/\/)?(?:[^@\/\n]+@)?(?:www\.)?([^:\/\n]+)`
)

func hostnamePattern(domainName string) string {
	escapedDomainName := strings.Replace(domainName, ".", "\\.", -1)
	return fmt.Sprintf(hostnamePatternFormat, escapedDomainName)
}

func itoi64(i int) *int64 {

	i64 := int64(i)
	return &i64
}
