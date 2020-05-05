package testsuite

import (
	"fmt"
)

func (t *TestSuite) getFunction(value string) *functionData {
	return &functionData{
		Body: fmt.Sprintf(`module.exports = { main: function(event, context) { return "%s" } }`, value),
		Deps: `{ "name": "hellowithoutdeps", "version": "0.0.1", "dependencies": { } }`,
	}
}

func (t *TestSuite) getUpdatedFunction(eventDataKey, eventDataValue, succesfullyGotEventMsg string) *functionData {
	return &functionData{
		// such a function tests simultaneously importing external lib, the fact that it was triggered (by using counter) and passing argument to function in event
		Body: fmt.Sprintf(`
const _ = require("lodash");

let counter = 0;
let eventCounter = 0;

module.exports = {
  main: function (event, context) {
    try {
      counter = _.add(counter, 1);
	  console.log(event.data)
      const eventData = event.data["%s"];
      if(eventData==="%s"){
      	eventCounter = 1;
      }

      if(eventCounter!==0){
		 console.log("eventCounter" + eventCounter, " Counter " + counter)
         return "%s"
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
`, eventDataKey, eventDataValue, succesfullyGotEventMsg),
		Deps:        `{ "name": "hellowithdeps", "version": "0.0.1", "dependencies": { "lodash": "^4.17.5" } }`,
		MaxReplicas: 2,
		MinReplicas: 1,
	}
}
