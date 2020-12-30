build: ## Build a version
	go build -v ./cmd/...

lint: ## Run linters
	golangci-lint run --no-config --issues-exit-code=0 --deadline=30m \
    --disable-all --enable=deadcode  --enable=gocyclo --enable=golint --enable=varcheck \
    --enable=structcheck --enable=maligned --enable=errcheck --enable=dupl --enable=ineffassign \
    --enable=interfacer --enable=unconvert --enable=goconst --enable=gosec --enable=megacheck

test:	## Run all the tests
	go test -v -race -timeout 30s ./...

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-17s\033[0m %s\n", $$1, $$2}'
