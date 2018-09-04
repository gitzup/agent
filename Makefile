ASSET_FILES = $(shell find ./api -type f)
ASSET_DIRS = $(shell find ./api -type d)
INTERNAL_SRC = $(shell find ./internal -type f -name '*.go')
TAG ?= dev

build: internal/assets.go agent

.PHONY: clean
clean:
	rm -vf agent

internal/assets.go: $(ASSET_FILES)
	$(GOPATH)/bin/go-bindata -o ./internal/assets.go -pkg internal -prefix api/ $(ASSET_DIRS)

agent: ./cmd/agent.go $(INTERNAL_SRC) $(ASSET_FILES)
	go build ./cmd/agent.go

.PHONY: docker
docker: agent
	docker build --tag gitzup/agent:$(TAG) --file ./build/Dockerfile .

.PHONY: docker
push-docker: docker
	docker push gitzup/agent:$(TAG)
