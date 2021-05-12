
const prometheusURL = "http://monitoring-prometheus.kyma-system:9090";
const namespace = "kyma-system";
const expectedAlertManagers = 1;
const expectedPrometheusInstances = 1;
const expectedKubeStateMetrics = 1;
const expectedGrafanaInstance = 1;

var HttpClient = function() {
    this.get = function(aUrl, aCallback) {
        var anHttpRequest = new XMLHttpRequest();
        anHttpRequest.onreadystatechange = function() { 
            if (anHttpRequest.readyState == 4 && anHttpRequest.status == 200)
                aCallback(anHttpRequest.responseText);
        }

        anHttpRequest.open( "GET", aUrl, true );            
        anHttpRequest.send( null );
    }
}
async function getAllPodsTest() {
    const { body } = await k8sCoreV1Api.listPodForAllNamespaces();
}



async function checkPrometheusRules() {
    await prometheusRules();
}

async function prometheusRules(retryInterval, promise) {

    promise = promise||new Promise();
    var timeoutMessage;

    const url = `${prometheusURL}/api/v1/rules`
    var client = new HttpClient();
    client.get(url, function(response) {
        let resp_array = JSON.parse( response );
        let allRulesAreHealthy = true
		let rulesGroups = resp_array.Data.Groups
		timeoutMessage = ""

        var i;
        for (i = 0; i < rulesGroups.length(); i++) {
            var j;
            for( j= 0; j < rulesGroups[i].length; j ++) {
                if ( rulesGroups[i][j] != "ok") {
                    allRulesAreHealthy = false
                    timeoutMessage += `Rule with name=${rulesGroups[i][j].Name} is not healthy\n`
                }
            }
        }
	
	if(allRulesAreHealthy) {
		promise.resolve(result);
	} else {
		setTimeout(function() {
			prometheusRules(Math.min(maxRetryInterval, retryInterval * 2), promise);
		}, retryInterval);
	}

    });
}