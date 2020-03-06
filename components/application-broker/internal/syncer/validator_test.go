package syncer

import (
	"io/ioutil"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestAppCRValidatorValidateSuccess(t *testing.T) {
	// given
	application := mustLoadCRFix("testdata/app-CR-valid.input.yaml")
	validator := &appCRValidator{}

	// when
	err := validator.Validate(&application)

	// then
	assert.NoError(t, err)
}

func TestAppCRValidatorValidateFailure(t *testing.T) {
	tests := map[string]struct {
		fixModifier func(*v1alpha1.Application)
		expErrMsg   []string
	}{
		"empty entries list": {
			fixModifier: func(app *v1alpha1.Application) {
				app.Spec.Services[0].Entries = []v1alpha1.Entry{}
			},
			expErrMsg: []string{"Service with id \"ac031e8c-9aa4-4cb7-8999-0d358726ffaa\" is invalid. Entries list cannot be empty"},
		},
		"missing GatewayUrl field": {
			fixModifier: func(app *v1alpha1.Application) {
				for i := range app.Spec.Services[0].Entries {
					app.Spec.Services[0].Entries[i].GatewayUrl = ""
				}
			},
			expErrMsg: []string{"GatewayUrl field is required for API type"},
		},
		"missing AccessLabel field": {
			fixModifier: func(app *v1alpha1.Application) {
				for i := range app.Spec.Services[0].Entries {
					app.Spec.Services[0].Entries[i].AccessLabel = ""
				}
			},
			expErrMsg: []string{"AccessLabel field is required for API type"},
		},
		"multiple API entries in one service": {
			fixModifier: func(app *v1alpha1.Application) {
				app.Spec.Services[0].Entries = []v1alpha1.Entry{
					{Type: "API", GatewayUrl: "test.svc.1", AccessLabel: "access.1"},
					{Type: "API", GatewayUrl: "test.svc.2", AccessLabel: "access.2"},
				}
			},
			expErrMsg: []string{"Service with id \"ac031e8c-9aa4-4cb7-8999-0d358726ffaa\" is invalid. Only one element with type API is allowed but found 2"},
		},
		"multiple Event entries in one service": {
			fixModifier: func(app *v1alpha1.Application) {
				app.Spec.Services[0].Entries = []v1alpha1.Entry{
					{Type: "Events"},
					{Type: "Events"},
				}
			},
			expErrMsg: []string{"Service with id \"ac031e8c-9aa4-4cb7-8999-0d358726ffaa\" is invalid. Only one element with type Event is allowed but found 2"},
		},
		"Labels filed does not contains required entries": {
			fixModifier: func(app *v1alpha1.Application) {
				app.Spec.Services[0].Labels = map[string]string{}
			},
			expErrMsg: []string{"Service with id \"ac031e8c-9aa4-4cb7-8999-0d358726ffaa\" is invalid. Labels field does not contains connected-app entry"},
		},
		"multiple validation errors": {
			fixModifier: func(app *v1alpha1.Application) {
				for i := range app.Spec.Services[0].Entries {
					app.Spec.Services[0].Entries[i].AccessLabel = ""
				}
				for i := range app.Spec.Services[0].Entries {
					app.Spec.Services[0].Entries[i].GatewayUrl = ""
				}
			},
			expErrMsg: []string{"GatewayUrl field is required for API type", "AccessLabel field is required for API type"},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			validator := &appCRValidator{}
			fixCR := mustModifyValidCR(tc.fixModifier)

			// when
			err := validator.Validate(fixCR)

			// then
			for _, msg := range tc.expErrMsg {
				assert.Contains(t, err.Error(), msg)
			}
		})
	}
}

func mustModifyValidCR(modifer func(app *v1alpha1.Application)) *v1alpha1.Application {
	fix := mustLoadCRFix("testdata/app-CR-valid.input.yaml")
	modifer(&fix)

	return &fix
}

func mustLoadCRFix(path string) v1alpha1.Application {
	in, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var application v1alpha1.Application
	err = yaml.Unmarshal(in, &application)
	if err != nil {
		panic(err)
	}

	return application
}
