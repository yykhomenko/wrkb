build: ## Build a version
	go build -v ./cmd/...

lint: ## Run linters
	golangci-lint run --no-config --issues-exit-code=0 --deadline=30s \
	--enable-all --disable=wsl --disable=nlreturn --disable=wrapcheck \
	--disable=forbidigo --disable=gofumpt

test:	## Run all the tests
	go test -v -race -timeout 30s ./...

run: ## Run wrkb
	./wrkb main http://127.0.0.1:8080

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-17s\033[0m %s\n", $$1, $$2}'
