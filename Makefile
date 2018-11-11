ASSET_FILES = $(shell find ./api -type f)
ASSET_DIRS = $(shell find ./api -type d)
SRC = $(shell find ./cmd ./internal ./pkg -type f -name '*.go')
TAG ?= dev

build: agent

.PHONY: clean
clean:
	rm -vf agent

pkg/assets/assets.go: $(ASSET_FILES)
	go-bindata -o pkg/assets/assets.go -pkg assets -prefix api/ $(ASSET_DIRS)

agent: ./main.go $(SRC) pkg/assets/assets.go
	go build -o agent ./main.go

.PHONY: docker
docker:
	docker build --tag gitzup/agent:$(TAG) --file ./build/Dockerfile .
	[[ "${PUSH}" == "false" ]] || docker push gitzup/agent:$(TAG)

.PHONY: latest
latest: docker
	docker tag gitzup/agent:$(TAG) gitzup/agent:latest
	[[ "${PUSH}" == "false" ]] || docker push gitzup/agent:latest
