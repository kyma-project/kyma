package runtimes

import (
	"fmt"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	corev1 "k8s.io/api/core/v1"
)

func BasicNodeJSFunction(msg string, rtm serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	return serverlessv1alpha2.FunctionSpec{
		Runtime: rtm,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       fmt.Sprintf(`module.exports = { main: function(event, context) { return "%s" } }`, msg),
				Dependencies: `{ "name": "hellobasic", "version": "0.0.1", "dependencies": {} }`,
			},
		},
		ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
			Function: &serverlessv1alpha2.ResourceRequirements{
				Profile: "M",
			},
			Build: &serverlessv1alpha2.ResourceRequirements{
				Profile: "normal",
			},
		},
	}
}

func BasicTracingNodeFunction(rtm serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	return serverlessv1alpha2.FunctionSpec{
		//TODO: add nodejs function with calling to ngix
		Runtime: rtm,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source: `const axios = require("axios")

module.exports = {
    main: async function (event, context) {

        axios.interceptors.response.use(res => {
            console.log("axios request headers", res.request._header)
            return res;
          }, error => Promise.reject(error));

        console.log("request headers",event.extensions.request.headers)
        let resp = await axios("https://swapi.dev/api/people/1");
        return resp.data;
    }
}`,
				Dependencies: `{
  "name": "sanitise-fn",
  "version": "0.0.1",
  "dependencies": {
    "axios":"0.26.1"
  }
}`,
			},
		},
	}
}

func BasicNodeJSFunctionWithCustomDependency(msg string, rtm serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	return serverlessv1alpha2.FunctionSpec{
		Runtime: rtm,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       fmt.Sprintf(`module.exports = { main: function(event, context) { return "%s" } }`, msg),
				Dependencies: `{ "name": "hellobasic", "version": "0.0.1", "dependencies": { "camelcase": "^7.0.0" } }`,
			},
		},
		ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
			Function: &serverlessv1alpha2.ResourceRequirements{
				Profile: "M",
			},
			Build: &serverlessv1alpha2.ResourceRequirements{
				Profile: "normal",
			},
		},
	}
}

func NodeJSFunctionWithEnvFromConfigMapAndSecret(configMapName, cmEnvKey, secretName, secretEnvKey string, rtm serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	mappedCmEnvKey := "CM_KEY"
	mappedSecretEnvKey := "SECRET_KEY"

	src := fmt.Sprintf(`module.exports = { main: function(event, context) { return process.env["%s"] + "-" + process.env["%s"]; } }`, mappedCmEnvKey, mappedSecretEnvKey)
	dpd := `{ "name": "hellowithconfigmapsecretenvs", "version": "0.0.1", "dependencies": { } }`

	return serverlessv1alpha2.FunctionSpec{
		Runtime: rtm,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       src,
				Dependencies: dpd,
			},
		},
		Env: []corev1.EnvVar{
			{
				Name: mappedCmEnvKey,
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: configMapName,
						},
						Key: cmEnvKey,
					},
				},
			},
			{
				Name: mappedSecretEnvKey,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: secretName,
						},
						Key: secretEnvKey,
					},
				},
			},
		},
		ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
			Function: &serverlessv1alpha2.ResourceRequirements{
				Profile: "M",
			},
			Build: &serverlessv1alpha2.ResourceRequirements{
				Profile: "normal",
			},
		},
	}
}
