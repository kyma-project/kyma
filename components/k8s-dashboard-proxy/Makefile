NAME = k8s-dashboard-proxy
VERSION = 0.1.0
REGISTRY = eu.gcr.io
.PHONY: all clean compile docker tag
all: clean compile docker tag
clean: 
	rm -rf bin/
	rm -rf docker/
compile:
	go build -o bin/k8s-dashboard-proxy ./cmd/reverseproxy
docker:
	mkdir -p docker/image
	cp Dockerfile docker/image/
	cp -R ./cmd/reverseproxy docker/image/
	docker build --build-arg version=$(VERSION) -t $(NAME):$(VERSION) --rm docker/image
	rm -rf docker
tag:
	docker tag $(NAME):$(VERSION) $(REGISTRY)/$(NAME):$(VERSION)
