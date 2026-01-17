build: ## Build version
	go build ./cmd/...

test:	## Run tests
	go test -timeout 10s ./...

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

bench_pico:
	go run ./... \
		-p=main \
		-best-json=pico.json \
		-compare \
	  -t=1 \
	  -n=10000 \
	  -X=POST \
	  -H 'Authorization: Bearer eyJ4NXQi' \
	  -d='{"msisdn": __RANDI64_380670000000_380679999999__}' \
	  http://127.0.0.1:8088/

bench_hashes:
	go run ./... \
		-p=hashes \
	  -t=1 \
	  http://127.0.0.1:8082/hashes/__RANDI64_380670000000_380679999999__

bench_hashes_kube:
	go run ./... \
		-p=hashes \
	  -t=1 \
	  http://127.0.0.1:8080/hashes/__RANDI64_380670000000_380679999999__

bench_sis_get:
	go run ./... \
		-p=sis \
		-t=1 \
	  -H 'Authorization: Bearer eyJ4NXQi' \
	  http://127.0.0.1:9001/subscribers/__RANDI64_380670000000_380719999999__

bench_sis_update:
	go run ./... \
		-p=main \
		-t=10 \
	  -c=12 \
	  -n=100000 \
	  -X=PUT \
	  -H 'Authorization: Bearer eyJ4NXQi' \
	  -d='{"billing_type": __RANDI64_0_2__, "language_type": __RANDI64_0_2__, "operator_type": __RANDI64_0_2__}' \
	  http://127.0.0.1:9001/subscribers/__SEQI64_380500000000_380509999999__

help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / \
  {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
