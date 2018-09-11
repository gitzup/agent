ASSET_FILES = $(shell find ./api -type f)
ASSET_DIRS = $(shell find ./api -type d)
INTERNAL_SRC = $(shell find ./cmd ./internal -type f -name '*.go')
TAG ?= dev

build: app

.PHONY: clean
clean:
	rm -vf agent

internal/assets.go: $(ASSET_FILES)
	$(GOPATH)/bin/go-bindata -o ./internal/assets.go -pkg internal -prefix api/ $(ASSET_DIRS)

app: ./main.go $(INTERNAL_SRC) $(ASSET_FILES)
	go build -o app ./main.go

.PHONY: docker
docker:
	docker build --tag gitzup/agent:$(TAG) --file ./build/Dockerfile .

.PHONY: docker
push-docker: docker
	docker push gitzup/agent:$(TAG)
	docker tag gitzup/agent:$(TAG) gitzup/agent:latest
	docker push gitzup/agent:latest
