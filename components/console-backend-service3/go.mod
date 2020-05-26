module github.com/kyma-project/kyma/components/console-backend-service3

go 1.13

require (
	github.com/99designs/gqlgen v0.11.3
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/websocket v1.4.0
	github.com/kyma-project/kyma v0.5.1-0.20200522083930-7e30b04a20f1
	github.com/opentracing/opentracing-go v1.1.0
	github.com/openzipkin-contrib/zipkin-go-opentracing v0.4.5
	github.com/pkg/errors v0.8.1
	github.com/rs/cors v1.6.0
	github.com/stretchr/testify v1.4.0
	github.com/vektah/gqlparser/v2 v2.0.1
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b // indirect
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/apiserver v0.17.4
	k8s.io/client-go v0.17.4
)
