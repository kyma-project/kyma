NAME = k8s-dashboard-proxy
VERSION = 0.1.0
REGISTRY = eu.gcr.io
.PHONY: all clean compile docker tag
all: clean compile docker tag
clean: 
	rm -rf docker/
compile:
	mkdir -p ./bin
	go build ./cmd/reverseproxy
	mv ./reverseproxy ./bin
docker:
	mkdir -p docker/image
	cp Dockerfile docker/image/
	cp -R cmd util docker/image/
	docker build --no-cache --build-arg version=$(VERSION) -t $(NAME):$(VERSION) --label version=${VERSION} --label component=${NAME} --rm docker/image
	rm -rf docker
tag:
	docker tag $(NAME):$(VERSION) $(REGISTRY)/$(NAME):$(VERSION)

vet:
	go list ./... | grep -v generated | xargs go vet
