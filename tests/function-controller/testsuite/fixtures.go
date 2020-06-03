package testsuite

import (
	"bytes"
	"errors"
	"fmt"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"text/template"
)

func (t *TestSuite) getFunction(value string) *functionData {
	return &functionData{
		Body: fmt.Sprintf(`module.exports = { main: function(event, context) { return "%s" } }`, value),
		Deps: `{ "name": "hellowithoutdeps", "version": "0.0.1", "dependencies": { } }`,
	}
}

func (t *TestSuite) getUpdatedFunction() *functionData {
	return &functionData{
		// such a function tests simultaneously importing external lib, the fact that it was triggered (by using counter) and passing argument to function in event
		Body:        getBodyString(),
		Deps:        `{ "name": "hellowithdeps", "version": "0.0.1", "dependencies": { "lodash": "^4.17.5" } }`,
		MaxReplicas: 2,
		MinReplicas: 1,
	}
}

func (t *TestSuite) checkDefaultedFunction(function *serverlessv1alpha1.Function) error {
	if function == nil {
		return errors.New("function can't be nil")
	}

	spec := function.Spec
	if spec.MinReplicas == nil {
		return errors.New("minReplicas equal nil")
	} else if spec.MaxReplicas == nil {
		return errors.New("maxReplicas equal nil")
	} else if spec.Resources.Requests.Memory().IsZero() || spec.Resources.Requests.Cpu().IsZero() {
		return errors.New("requests equal zero")
	} else if spec.Resources.Limits.Memory().IsZero() || spec.Resources.Limits.Cpu().IsZero() {
		return errors.New("limits equal zero")
	}
	return nil
}

func getBodyString() string {
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
		TestData:      testDataKey,
		SbuEventValue: redisEnvPing,
		RedisPortEnv:  fmt.Sprintf("%s%s", redisEnvPrefix, "PORT"),
		EventPing:     eventPing,
		EventAnswer:   gotEventMsg,
	}

	var tpl bytes.Buffer
	err := t.Execute(&tpl, data)
	if err != nil {
		panic(err)
	}
	return tpl.String()
}
