package runtimes

import (
	"bytes"
	"fmt"
	"html/template"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	corev1 "k8s.io/api/core/v1"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/function"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite"
)

func BasicNodeJSFunction(msg string, rtm serverlessv1alpha1.Runtime) *function.FunctionData {
	return &function.FunctionData{
		Body:        fmt.Sprintf(`module.exports = { main: function(event, context) { return "%s" } }`, msg),
		Deps:        `{ "name": "hellowithoutdeps", "version": "0.0.1", "dependencies": { } }`,
		MaxReplicas: 2,
		MinReplicas: 1,
		Runtime:     rtm,
	}
}

func NodeJSFunctionWithEnvFromConfigMapAndSecret(configMapName, cmEnvKey, secretName, secretEnvKey string, rtm serverlessv1alpha1.Runtime) *function.FunctionData {
	mappedCmEnvKey := "CM_KEY"
	mappedSecretEnvKey := "SECRET_KEY"

	return &function.FunctionData{
		Body:        fmt.Sprintf(`module.exports = { main: function(event, context) { return process.env["%s"] + "-" + process.env["%s"]; } }`, mappedCmEnvKey, mappedSecretEnvKey),
		Deps:        `{ "name": "hellowithconfigmapsecretenvs", "version": "0.0.1", "dependencies": { } }`,
		MaxReplicas: 1,
		MinReplicas: 1,
		Runtime:     rtm,
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
				}},
			{
				Name: mappedSecretEnvKey,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: secretName,
						},
						Key: secretEnvKey,
					},
				}},
		},
	}
}

func GetUpdatedNodeJSFunction() *function.FunctionData {
	return &function.FunctionData{
		// such a Function tests simultaneously importing external lib, the fact that it was triggered (by using counter) and passing argument to Function in event
		Body:        getNodeJSUpdatedBody(),
		Deps:        `{ "name": "hellowithdeps", "version": "0.0.1", "dependencies": { "lodash": "^4.17.5" } }`,
		MaxReplicas: 2,
		MinReplicas: 1,
	}
}

func getNodeJSUpdatedBody() string {
	t := template.Must(template.New("body").Parse(
		`
const _ = require("lodash");

let counter = 0;
let eventCounter = 0;

module.exports = {
  main: function (event, context) {
    try {
      counter = _.add(counter, 1);
	  console.log(event.data)
      const eventData = event.data["{{ .TestData }}"];
  
      if(eventData==="{{ .SbuEventValue }}"){
        const ret = process.env["{{ .RedisPortEnv }}"]
	  	console.log("Redis port: " + ret);
		return "Redis port: " + ret;	
      }    

      if(eventData==="{{ .EventPing }}"){
      	eventCounter = 1;
      }

      if(eventCounter!==0){
		 console.log("eventCounter" + eventCounter, " Counter " + counter)
         return "{{ .EventAnswer }}"
      }

      const answer = "Hello " + eventData + " world " + counter;
      console.log(answer);
      return answer;
    } catch (err) {
	  console.error(event);
      console.error(context);
	  console.error(err);
      return "Failed to parse event. Counter value: " + counter;
    }
  }
}
`))

	data := struct {
		TestData      string
		SbuEventValue string
		RedisPortEnv  string
		EventPing     string
		EventAnswer   string
	}{
		TestData:      testsuite.TestDataKey,
		SbuEventValue: testsuite.RedisEnvPing,
		RedisPortEnv:  fmt.Sprintf("%s%s", testsuite.RedisEnvPrefix, "PORT"),
		EventPing:     testsuite.EventPing,
		EventAnswer:   testsuite.GotEventMsg,
	}

	var tpl bytes.Buffer
	err := t.Execute(&tpl, data)
	if err != nil {
		panic(err)
	}
	return tpl.String()
}
