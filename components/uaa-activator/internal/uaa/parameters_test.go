package uaa

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	clusterDomain = "uaa-test.kyma-dev.shoot.canary.k8s-hana.ondemand.com"
	randomSuffix  = "_lxfaa"
)

func TestParametersBuilder_Generate(t *testing.T) {
	t.Run("upgrade parameters for the old instance version", func(t *testing.T) {
		// Given
		pb := NewParametersBuilder(Config{
			DeveloperGroup:      "runtimeDeveloper",
			DeveloperRole:       "KymaRuntimeDeveloper",
			NamespaceAdminGroup: "runtimeNamespaceAdmin",
			NamespaceAdminRole:  "KymaRuntimeNamespaceAdmin",
		}, clusterDomain)

		// When
		parameters, err := pb.Generate(fixOldServiceInstance())
		require.NoError(t, err)

		// Then
		schema := &Schema{}
		err = json.Unmarshal(parameters, schema)
		require.NoError(t, err)

		require.Equal(t, schema.Xsappname, strings.ReplaceAll(clusterDomain, ".", "_"))
		parametersAreEqual(t, *schema)
	})

	t.Run("upgrade parameters for the new instance version", func(t *testing.T) {
		// Given
		pb := NewParametersBuilder(Config{
			DeveloperGroup:      "runtimeDeveloper",
			DeveloperRole:       "KymaRuntimeDeveloper",
			NamespaceAdminGroup: "runtimeNamespaceAdmin",
			NamespaceAdminRole:  "KymaRuntimeNamespaceAdmin",
		}, clusterDomain)

		// When
		parameters, err := pb.Generate(fixServiceInstance())
		require.NoError(t, err)

		// Then
		schema := &Schema{}
		err = json.Unmarshal(parameters, schema)
		require.NoError(t, err)

		require.Equal(t, schema.Xsappname, fmt.Sprintf("%s%s", strings.ReplaceAll(clusterDomain, ".", "_"), randomSuffix))
		parametersAreEqual(t, *schema)
	})

	t.Run("parameters for the new instance", func(t *testing.T) {
		// Given
		pb := NewParametersBuilder(Config{
			DeveloperGroup:      "runtimeDeveloper",
			DeveloperRole:       "KymaRuntimeDeveloper",
			NamespaceAdminGroup: "runtimeNamespaceAdmin",
			NamespaceAdminRole:  "KymaRuntimeNamespaceAdmin",
		}, clusterDomain)

		// When
		parameters, err := pb.Generate(nil)
		require.NoError(t, err)

		// Then
		schema := &Schema{}
		err = json.Unmarshal(parameters, schema)
		require.NoError(t, err)

		require.Regexp(t, "^(uaa-test_kyma-dev_shoot_canary_k8s-hana_ondemand_com)\\_[a-z]{5,5}$", schema.Xsappname)
		parametersAreEqual(t, *schema)
	})
}

func parametersAreEqual(t *testing.T, schema Schema) {
	roles := []string{"KymaRuntimeDeveloper", "KymaRuntimeNamespaceAdmin"}
	rolesWithSuffix := []string{"KymaRuntimeDeveloper___uaa_test", "KymaRuntimeNamespaceAdmin___uaa_test"}
	rolesWithName := []string{"$XSAPPNAME.KymaRuntimeDeveloper", "$XSAPPNAME.KymaRuntimeNamespaceAdmin"}
	groups := []string{"$XSAPPNAME.runtimeDeveloper", "$XSAPPNAME.runtimeNamespaceAdmin"}
	scopesGroup := []string{"$XSAPPNAME.email", "$XSAPPNAME.runtimeDeveloper", "$XSAPPNAME.runtimeNamespaceAdmin"}

	require.Equal(t, schema.TenantMode, "shared")

	require.Len(t, schema.Scopes, 3)
	names := make([]string, 0)
	for _, s := range schema.Scopes {
		names = append(names, s.Name)
	}
	require.ElementsMatch(t, scopesGroup, names)

	require.Len(t, schema.Authorities, 1)
	require.Equal(t, schema.Authorities[0], "$ACCEPT_GRANTED_AUTHORITIES")

	require.Len(t, schema.Oauth2Configuration.RedirectUris, 1)
	require.Equal(t, schema.Oauth2Configuration.RedirectUris[0], fmt.Sprintf("https://dex.%s/callback", clusterDomain))

	require.Len(t, schema.RoleTemplates, 2)
	require.Contains(t, roles, schema.RoleTemplates[0].Name)
	require.Contains(t, roles, schema.RoleTemplates[1].Name)
	require.Len(t, schema.RoleTemplates[0].ScopeReferences, 1)
	require.Len(t, schema.RoleTemplates[1].ScopeReferences, 1)
	require.Contains(t, groups, schema.RoleTemplates[0].ScopeReferences[0])
	require.Contains(t, groups, schema.RoleTemplates[1].ScopeReferences[0])

	require.Len(t, schema.RoleCollections, 2)
	require.Contains(t, rolesWithSuffix, schema.RoleCollections[0].Name)
	require.Contains(t, rolesWithSuffix, schema.RoleCollections[1].Name)
	require.Len(t, schema.RoleCollections[0].RoleTemplateReference, 1)
	require.Len(t, schema.RoleCollections[1].RoleTemplateReference, 1)
	require.Contains(t, rolesWithName, schema.RoleCollections[0].RoleTemplateReference[0])
	require.Contains(t, rolesWithName, schema.RoleCollections[1].RoleTemplateReference[0])
}

func parameters() string {
	return fmt.Sprintf(
		`{
  "authorities": [
    "$ACCEPT_GRANTED_AUTHORITIES"
  ],
  "oauth2-configuration": {
    "redirect-uris": [
      "https://dex.%s/callback"
    ],
    "system-attributes": [
      "groups",
      "rolecollections"
    ]
  },
  "role-templates": [
    {
      "description": "Runtime developer access to all managed resources",
      "name": "KymaRuntimeNamespaceDeveloper",
      "scope-references": [
        "$XSAPPNAME.runtimeDeveloper"
      ]
    },
    {
      "description": "Runtime admin access to all managed resources",
      "name": "KymaRuntimeNamespaceAdmin",
      "scope-references": [
        "$XSAPPNAME.runtimeNamespaceAdmin"
      ]
    }
  ],
  "scopes": [
    {
      "description": "get user email",
      "name": "$XSAPPNAME.email"
    },
    {
      "description": "Runtime developer access to all managed resources",
      "name": "$XSAPPNAME.runtimeDeveloper"
    },
    {
      "description": "Runtime admin access to all managed resources",
      "name": "$XSAPPNAME.runtimeNamespaceAdmin"
    }
  ],
  "tenant-mode": "shared",
  "xsappname": "%s%s"
}`, clusterDomain, strings.ReplaceAll(clusterDomain, ".", "_"), randomSuffix)
}

func fixServiceInstance() *v1beta1.ServiceInstance {
	return &v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ClusterServiceClassName: "test",
				ClusterServicePlanName:  "test",
			},
			Parameters: &runtime.RawExtension{Raw: []byte(parameters())},
		},
	}
}

func fixOldServiceInstance() *v1beta1.ServiceInstance {
	return &v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ClusterServiceClassName: "test",
				ClusterServicePlanName:  "test",
			},
			ParametersFrom: []v1beta1.ParametersFromSource{
				{
					SecretKeyRef: &v1beta1.SecretKeyReference{
						Name: "secret-name",
						Key:  "secret-key",
					},
				},
			},
		},
	}
}
