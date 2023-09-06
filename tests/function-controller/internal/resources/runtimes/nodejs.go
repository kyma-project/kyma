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

func BasicTracingNodeFunction(rtm serverlessv1alpha2.Runtime, externalSvcURL string) serverlessv1alpha2.FunctionSpec {
	dpd := `{
  "name": "sanitise-fn",
  "version": "0.0.1",
  "dependencies": {
    "axios":"0.26.1"
  }
}`
	src := fmt.Sprintf(`const axios = require("axios")


module.exports = {
    main: async function (event, context) {
        let resp = await axios("%s",{timeout: 1000});
        let interceptedHeaders = resp.request._header
        let tracingHeaders = getTracingHeaders(interceptedHeaders)
        return tracingHeaders
    }
}

function getTracingHeaders(textHeaders) {
    tracingHeaders = textHeaders.split('\n')
        .filter(val => {
            let out = val.split(":")
            return out.length === 2;
        })
        .map(item => {
            let stringHeader = item.split(":")
            return {
                key: stringHeader[0],
                value: stringHeader[1]
            }
        })
        .filter(item => {
            return item.key.startsWith("x-b3") || item.key.startsWith("traceparent");
        })
        .map(val => {
            return {
                [val.key]: val.value
            }
        })
        .reduce((prev, current) => {
            return Object.assign(prev, current)
        })
    return tracingHeaders
}`, externalSvcURL)
	return serverlessv1alpha2.FunctionSpec{
		Runtime: rtm,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       src,
				Dependencies: dpd,
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

func NodeJSFunctionWithCloudEvent(rtm serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	src := `const process = require("process");
const axios = require('axios');
const {response} = require("express");

let cloudevent = {}

send_check_event_type = "send-check"

module.exports = {
    main: async function (event, context) {
        switch (event.extensions.request.method) {
            case "POST":
                return handlePost(event)
            case "GET":
                return handleGet(event.extensions.request)
            default:
                event.extensions.response.statusCode = 405
                return ""
        }
    }
}

function handlePost(event) {
    if (!Object.keys(event).includes("ce-type")) {
        event.emitCloudEvent(send_check_event_type, 'function', event.data, {'eventtypeversion': 'v1alpha2'})
        return ""
    }
    Object.keys(event).filter((val) => {
        return val.startsWith("ce-")
    }).forEach((item) => {
        cloudevent[item] = event[item]
    })
    cloudevent.data = event.data
    return ""
}

async function handleGet(req) {
    if (req.query.type === 'send-check') {
        let data = {}
        let publisherProxy = process.env.PUBLISHER_PROXY_ADDRESS
        await axios.get(publisherProxy, {
            params: {
                type: req.query.type
            }
        }).then((res) => {
            data = res.data
        })
        return data
    }
    return cloudevent
}
`
	return serverlessv1alpha2.FunctionSpec{
		Runtime: rtm,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       src,
				Dependencies: `{ "name": "cloudevent", "version": "0.0.1", "dependencies": {} }`,
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
