---
title: Develop a service locally without using Docker
type: Tutorials
---

You can develop services in the local Kyma installation without extensive Docker knowledge or a need to build and publish a Docker image. The `minikube mount` feature allows you to mount a directory from your local disk into the local Kubernetes cluster.

This tutorial shows how to use this feature, using the service example implemented in Golang.

## Prerequisites

Install [Golang](https://golang.org/dl/).

## Steps

### Install the example on your local machine

1. Install the example:
```shell
go get -insecure github.com/kyma-project/examples/http-db-service
```
2. Navigate to installed example and the `http-db-service` folder inside it:
```shell
cd ~/go/src/github.com/kyma-project/examples/http-db-service
```
3. Build the executable to run the application:
```shell
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
```

### Mount the example directory into Minikube

For this step, you need a running local Kyma instance. Read [this](#installation-install-kyma-locally-from-the-release) document to learn how to install Kyma locally.

1. Open the terminal window. Do not close it until the development finishes.
2. Mount your local drive into Minikube:
```shell
# Use the following pattern:
minikube mount {LOCAL_DIR_PATH}:{CLUSTER_DIR_PATH}`
# To follow this guide, call:
minikube mount ~/go/src/github.com/kyma-project/examples/http-db-service:/go/src/github.com/kyma-project/examples/http-db-service
```

See the example and expected result:
```shell
# Terminal 1
$ minikube mount ~/go/src/github.com/kyma-project/examples/http-db-service:/go/src/github.com/kyma-project/examples/http-db-service

Mounting /Users/{USERNAME}/go/src/github.com/kyma-project/examples/http-db-service into /go/src/github.com/kyma-project/examples/http-db-service on the minikube VM
This daemon process must stay alive for the mount to still be accessible...
ufs starting
```

### Run your local service inside Minikube

1. Create Pod that uses the base Golang image to run your executable located on your local machine:
```shell
# Terminal 2
kubectl run mydevpod --image=golang:1.9.2-alpine --restart=Never -n stage --overrides='
{
   "spec":{
      "containers":[
         {
            "name":"mydevpod",
            "image":"golang:1.9.2-alpine",
            "command": ["./main"],
            "workingDir":"/go/src/github.com/kyma-project/examples/http-db-service",
            "volumeMounts":[
               {
                  "mountPath":"/go/src/github.com/kyma-project/examples/http-db-service",
                  "name":"local-disk-mount"
               }
            ]
         }
      ],
      "volumes":[
         {
            "name":"local-disk-mount",
            "hostPath":{
               "path":"/go/src/github.com/kyma-project/examples/http-db-service"
            }
         }
      ]
   }
}
'
```
2. Expose the Pod as a service from Minikube to verify it:
```shell
kubectl expose pod mydevpod --name=mypodservice --port=8017 --type=NodePort -n stage
```
3. Check the Minikube IP address and Port, and use them to access your service.
```shell
# Get the IP address.
minikube ip
# See the example result: 192.168.64.44
# Check the Port.
kubectl get services -n stage
# See the example result: mypodservice  NodePort 10.104.164.115  <none>  8017:32226/TCP  5m
```
4. Call the service from your terminal.
```shell
curl {minikube ip}:{port}/orders -v
# See the example: curl http://192.168.64.44:32226/orders -v
# The command returns an empty array.
```

### Modify the code locally and see the results immediately in Minikube

1. Edit the `main.go` file by adding a new `test` endpoint to the `startService` function
```go
router.HandleFunc("/test", func (w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("test"))
})
```
2. Build a new executable to run the application inside Minikube:
```shell
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
```
3. Replace the existing Pod with the new version:
```shell
kubectl get pod mydevpod -n stage -o yaml | kubectl replace --force -f -
```
4. Call the new `test` endpoint of the service from your terminal. The command returns the `Test` string:
```shell
curl http://192.168.64.44:32226/test -v
```
