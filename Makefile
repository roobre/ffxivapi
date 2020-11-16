DOCKER_TAG ?= "roobre/ffxivapi:latest"
bin := ffxivapi

ffixvapi: go.mod go.sum
	go mod download
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o ${bin} ./http/cmd
	chmod +rx ffxivapi

docker:
	docker build -t ${DOCKER_TAG} .

clean:
	rm ${bin}

.PHONY: docker clean
