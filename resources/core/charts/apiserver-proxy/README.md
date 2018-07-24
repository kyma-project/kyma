```
_  __     _                          _                       _____ _____    _____                            _____                     
| |/ /    | |                        | |                /\   |  __ \_   _|  / ____|                          |  __ \                    
| ' /_   _| |__   ___ _ __ _ __   ___| |_ ___  ___     /  \  | |__) || |   | (___   ___ _ ____   _____ _ __  | |__) | __ _____  ___   _
|  <| | | | '_ \ / _ \ '__| '_ \ / _ \ __/ _ \/ __|   / /\ \ |  ___/ | |    \___ \ / _ \ '__\ \ / / _ \ '__| |  ___/ '__/ _ \ \/ / | | |
| . \ |_| | |_) |  __/ |  | | | |  __/ ||  __/\__ \  / ____ \| |    _| |_   ____) |  __/ |   \ V /  __/ |    | |   | | | (_) >  <| |_| |
|_|\_\__,_|_.__/ \___|_|  |_| |_|\___|\__\___||___/ /_/    \_\_|   |_____| |_____/ \___|_|    \_/ \___|_|    |_|   |_|  \___/_/\_\\__, |
                                                                                                                                  __/ |
                                                                                                                                 |___/
```

## Overview

This API Server Proxy is an [Nginx](https://nginx.org/en/)-based, transparent proxy for the Kubernetes API. It is exposed for the external communication.


## Details

Kyma requires all APIs, including those provided by the Kubernetes API server, to be exposed in a consistent manner through Istio.

To expose an API through Istio, all of the Pods that run the service containers must contain an Envoy sidecar. You need an additional proxy, as you cannot inject an Envoy sidecar directly into the Kubernetes API server. As a workaround, deploy Nginx as a proxy for the Kubernetes API server. Istio injects an Envoy sidecar into the Pods that run Nginx.

Installing the Helm chart creates an ingress, which exposes the API server under the `apiserver` subdomain in the configured domain.
