
IMAGE?=nwik/harbor-operator
IMAGE_TAG?=latest

all:
	operator-sdk build $(IMAGE):$(IMAGE_TAG)

build:
	operator-sdk build nwik/harbor-operator:latest

k8s:
	operator-sdk generate k8s

push:
	docker push $(IMAGE):$(IMAGE_TAG)
