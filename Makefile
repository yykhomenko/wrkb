build: ## Build version
	go build ./cmd/...

test:	## Run tests
	go test -race -timeout 10s ./...

bench:
	go test -bench=. -benchmem ./...

run: ## Run version
	go run ./cmd/... main http://127.0.0.1

start: ## Run version
	./wrkb main http://127.0.0.1

install: ## Install version
	make build
	make test
	go install ./cmd/...

clean: ## Clean project
	rm -f wrkb

help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / \
  {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
