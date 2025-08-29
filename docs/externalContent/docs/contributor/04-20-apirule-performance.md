# APIRule Performance Tests

Tests were performed on Gardener GCP cluster with the help of `k6s` on Kyma `production` profile. Requests were made from inside the cluster with HPA disabled for the sake of having identical environment on all runs. The tests were made using 500 Virtual Users making constant requests for 1 minute.

## Performance of Deployments Without Istio Sidecar

|Handler type|No. of successfull calls|No. of failed calls|Data recieved by server [MB]|Transfer speed (recieving) [kB/s]|Data sent by server [MB]|Transfer speed (sending) [kB/s]|Median request duration [ms]|
|---|---|---|---|---|---|---|---|
|Allow|291993|13|34|569|11|180|91.69|
|Noop (Ory)|74779|0|11|183|3.0|49|328.17|
|OAuth2|24198|4|5.5|89|6.3|104|861.75|
|Ory JWT|66775|5|10|169|17|282|388.36|

## Performance of Deployments with Istio Sidecar

|Handler type|No. of successfull calls|No. of failed calls|Data recieved by server [MB]|Transfer speed (recieving) [kB/s]|Data sent by server [MB]|Transfer speed (sending) [kB/s]|Median request duration [ms]|
|---|---|---|---|---|---|---|---|
|Allow|242342|0|29|481|9.0|150|110.88|
|Noop (Ory)|75327|0|11|185|3.0|50|335.88|
|OAuth2|22295|99|5.3|87|5.9|97|1210.00|
|Istio JWT|208437|0|25|419|52|861|126.45|
|Ory JWT|63262|0|9.9|161|16|266|434.83|
