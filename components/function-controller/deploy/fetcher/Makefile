

build:
	docker build -t fetcher:stable -f Dockerfile .

k3d-push:
	k3d image import fetcher:stable --cluster kyma

local: build k3d-push
