NAME = kyma-project/event-bus-e2e-subscriber
VERSION = 0.1.0
REGISTRY = eu.gcr.io

.PHONY: all build docker-build tag push

all: clean build docker-build tag

clean:
	rm -rf bin/
	rm -rf docker/

build: vet compile

compile:
	go build -o bin/e2e-subscriber

docker-build:
	./dockerBuild.sh  $(NAME) $(VERSION)

tag:
	docker tag $(NAME):$(VERSION) $(REGISTRY)/$(NAME):$(VERSION)

push:
	docker push $(REGISTRY)/$(NAME):$(VERSION)

vet:
	go list ../... | xargs go vet
