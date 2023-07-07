build: ## Build version
	go build ./cmd/...

test:	## Run tests
	go test -race -timeout 10s ./...

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
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-17s\033[0m %s\n", $$1, $$2}'
