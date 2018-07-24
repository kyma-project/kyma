NAME = kyma-project/event-bus-sub-validator
VERSION = 0.1.0
REGISTRY = eu.gcr.io

.PHONY: all clean build tag push

all: clean build docker-build tag

clean:
	rm -rf bin/
	rm -rf docker/

build: vet compile

compile:
	go build -o bin/event-bus-sub-validator 

docker-build:
	mkdir -p docker/image
	cp Dockerfile docker/image/
	cp -R main.go application docker/image/
	cp -R ../../generated docker/image/
	cp -R ../../internal docker/image/
	cp -R ../../vendor docker/image/
	cp -R ../../api docker/image/
	docker build --build-arg version=$(VERSION) -t $(NAME):$(VERSION) --rm docker/image
	rm -rf docker

tag:
	docker tag $(NAME):$(VERSION) $(REGISTRY)/$(NAME):$(VERSION)

push:
	docker push $(REGISTRY)/$(NAME):$(VERSION)

vet:
	go list ../../... | grep -v generated | xargs go vet

