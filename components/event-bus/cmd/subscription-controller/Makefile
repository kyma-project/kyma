
NAME = kyma-project/subscription-controller
VERSION = 0.1.0
COMPONENT = event-bus
REGISTRY = eu.gcr.io

all: clean compile docker-build tag

clean:
	rm -rf bin/
	rm -rf docker/
	
# Run tests
test: generate fmt vet
	go test ./cmd/... -coverprofile cover.out

# Build manager binary
compile: generate fmt vet
	go build -o bin/subscription-controller github.com/kyma-project/kyma/components/event-bus/cmd/subscription-controller/cmd/manager

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./cmd/manager/main.go

# Run go fmt against code
fmt:
	go fmt ./cmd/...

# Run go vet against code
vet:
	go vet ./cmd/...

# Generate code
generate:
	go generate ./cmd/...

# Build the docker image
docker-build: 
	./dockerBuild.sh $(NAME) $(VERSION) $(COMPONENT)

# Tag the docker image
tag:
	docker tag $(NAME):$(VERSION) $(REGISTRY)/$(NAME):$(VERSION)
