module github.com/kyma-project/kyma/components/connector-service

go 1.16

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.0.0
	github.com/prometheus/client_model v0.1.0 // indirect
	github.com/prometheus/common v0.7.0 // indirect
	github.com/prometheus/procfs v0.0.8 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/tidwall/gjson v1.6.7
	golang.org/x/net v0.0.0-20191204025024-5ee1b9f4859a // indirect
	golang.org/x/oauth2 v0.0.0-20191202225959-858c2ad4c8b6 // indirect
	golang.org/x/sys v0.0.0-20191204072324-ce4227a45e2e // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	k8s.io/api v0.18.15
	k8s.io/apimachinery v0.18.15
	k8s.io/client-go v0.18.15
)

replace (
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.8.0
	github.com/tidwall/gjson => github.com/tidwall/gjson v1.6.7

	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad 
	golang.org/x/text => golang.org/x/text v0.3.3
)
