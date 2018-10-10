NAME = kyma-project/event-bus-push
VERSION = 0.1.0
COMPONENT = event-bus
REGISTRY = eu.gcr.io

.PHONY: all clean compile docker-build tag

all: clean compile docker-build tag

clean:
	rm -rf bin/
	rm -rf docker/

compile:
	go build -o bin/event-bus-push

docker-build:
	./dockerBuild.sh $(NAME) $(VERSION) $(COMPONENT)

tag:
	docker tag $(NAME):$(VERSION) $(REGISTRY)/$(NAME):$(VERSION)
