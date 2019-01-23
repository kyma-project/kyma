
NAME = kyma-project/event-bus-subscription-controller-knative
VERSION = 0.1.0
COMPONENT = event-bus
REGISTRY = eu.gcr.io

all: clean compile docker-build tag

clean:
	rm -rf bin/
	rm -rf docker/
	
# Run tests
test: generate fmt vet
	go test ./pkg/... ./cmd/... -coverprofile cover.out

# Build manager binary
compile: generate fmt vet
	go build -o bin/event-bus-subscription-controller-knative github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-subscription-controller-knative/cmd/manager

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./cmd/manager/main.go

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

# Generate code
generate:
	go generate ./pkg/... ./cmd/...

# Build the docker image
docker-build: 
	./dockerBuild.sh $(NAME) $(VERSION) $(COMPONENT)

# Tag the docker image
tag:
	docker tag $(NAME):$(VERSION) $(REGISTRY)/$(NAME):$(VERSION)
