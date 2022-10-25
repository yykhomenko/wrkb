build: ## Build version
	go build ./cmd/...

test:	## Run all tests
	go test -race -timeout 10s ./...

run: ## Run version
	./wrkb main http://127.0.0.1

clean:
	rm -f wrkb

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-17s\033[0m %s\n", $$1, $$2}'
