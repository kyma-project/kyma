package ybind_test

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/ybind"
	"github.com/renstrom/dedent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// Policy #1: When a key exists in multiple sources defined by `credentialFrom` section, then the value associated with the last source will take precedence
func TestResolvePolicyNo1(t *testing.T) {
	const (
		configName, secretName = "all-cfg-map-test-redis", "all-secret-test-redis"
		keyNo1, keyNo2, keyNo3 = "keyNo1", "keyNo2", "keyNo3"
		namespace              = "test-ns"
	)

	type given struct {
		configData configMapData
		secretData secretData
		bindYAML   string
	}
	type expected struct {
		credentials internal.InstanceCredentials
	}
	for tn, tc := range map[string]struct {
		given
		expected
	}{
		"secret overrides configMap values": {
			given: given{
				configData: configMapData{keyNo1: "key_1_cfg_val", keyNo2: "key_2_cfg_val"},
				secretData: secretData{keyNo1: []byte("key_1_secret_val"), keyNo3: []byte("key_3_secret_val")},
				bindYAML: dedent.Dedent(`
                      credentialFrom:
                        - configMapRef:
                           name: ` + configName + `
                        - secretRef:
                           name: ` + secretName),
			},
			expected: expected{
				credentials: internal.InstanceCredentials{
					keyNo1: "key_1_secret_val",
					keyNo2: "key_2_cfg_val",
					keyNo3: "key_3_secret_val",
				},
			},
		},
		"configMap overrides secret values": {
			given: given{
				configData: configMapData{keyNo1: "key_1_cfg_val", keyNo2: "key_2_cfg_val"},
				secretData: secretData{keyNo1: []byte("key_1_secret_val"), keyNo3: []byte("key_3_secret_val")},
				bindYAML: dedent.Dedent(`
		             credentialFrom:
		               - secretRef:
		                   name: ` + secretName + `
		               - configMapRef:
		                  name: ` + configName),
			},
			expected: expected{
				credentials: internal.InstanceCredentials{
					keyNo1: "key_1_cfg_val",
					keyNo2: "key_2_cfg_val",
					keyNo3: "key_3_secret_val",
				},
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			ts := newResolverTestSuit()
			var (
				configMap  = ts.configMap(namespace, "all-cfg-map-test-redis", tc.given.configData)
				secret     = ts.secret(namespace, "all-secret-test-redis", tc.given.secretData)
				fakeClient = fake.NewSimpleClientset(&configMap, &secret)
				resolver   = ybind.NewResolver(fakeClient.CoreV1())
			)

			// when
			out, err := resolver.Resolve(ybind.RenderedBindYAML(tc.given.bindYAML), internal.Namespace(namespace))

			// then
			require.NoError(t, err)
			assert.EqualValues(t, tc.expected.credentials, out.Credentials)
			assert.Len(t, fakeClient.Actions(), 2)
		})
	}
}

// Policy #2: When you duplicate a key in `credential` section then error will be returned
func TestResolvePolicyNo2(t *testing.T) {
	// given
	const (
		keyNo1    = "keyNo1"
		namespace = "test-ns"
	)

	bindYAML := dedent.Dedent(`
       credential:
        - name: ` + keyNo1 + `
          value: duplicated-value
        - name: ` + keyNo1 + `
          value: duplicated-value`)
	resolver := ybind.NewResolver(nil)

	// when
	out, err := resolver.Resolve(ybind.RenderedBindYAML(bindYAML), internal.Namespace(namespace))

	// then
	assert.EqualError(t, err, fmt.Sprintf("conflict: found credentials with the same name %q", keyNo1))
	assert.Nil(t, out)
}

// Policy #3: Values defined by `credentialFrom` section will be overridden by values from `credential` section if keys will be duplicated
func TestResolvePolicyNo3(t *testing.T) {
	// given
	const (
		configName, secretName = "all-cfg-map-test-redis", "all-secret-test-redis"
		keyNo1, keyNo2, keyNo3 = "keyNo1", "keyNo2", "keyNo3"
		namespace              = "test-ns"
	)

	ts := newResolverTestSuit()
	var (
		configData = configMapData{keyNo1: "key_1_cfg_val", keyNo2: "key_2_cfg_val"}
		secretData = secretData{keyNo1: []byte("key_1_secret_val"), keyNo3: []byte("key_3_secret_val")}
		bindYAML   = dedent.Dedent(`
          credential:
            - name: ` + keyNo1 + `
              value: override-value
          credentialFrom:
            - configMapRef:
                name: ` + configName + `
            - secretRef:
                name: ` + secretName)
		credentials = internal.InstanceCredentials{
			keyNo1: "override-value",
			keyNo2: "key_2_cfg_val",
			keyNo3: "key_3_secret_val",
		}
		configMap  = ts.configMap(namespace, "all-cfg-map-test-redis", configData)
		secret     = ts.secret(namespace, "all-secret-test-redis", secretData)
		fakeClient = fake.NewSimpleClientset(&configMap, &secret)
		resolver   = ybind.NewResolver(fakeClient.CoreV1())
	)

	// when
	out, err := resolver.Resolve(ybind.RenderedBindYAML(bindYAML), internal.Namespace(namespace))

	// then
	require.NoError(t, err)
	assert.EqualValues(t, credentials, out.Credentials)
	assert.Len(t, fakeClient.Actions(), 2)
}

type resolverServiceTestSuit struct{}

func newResolverTestSuit() *resolverServiceTestSuit {
	return &resolverServiceTestSuit{}
}

type configMapData map[string]string

func (*resolverServiceTestSuit) configMap(namespace, name string, data configMapData) v1.ConfigMap {
	return v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Data: data,
	}
}

type secretData map[string][]byte

func (*resolverServiceTestSuit) secret(namespace, name string, data secretData) v1.Secret {
	return v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Data: data,
	}
}
