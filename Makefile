DOCKER_TAG ?= "roobre/ffxivapi:latest"

ffixvapi:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o ffxivapi ./http/cmd

docker:
	docker build -t ${DOCKER_TAG} .

.PHONY: docker
