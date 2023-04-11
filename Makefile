ALL:
	go build ./cmd/udp-spoof/...

freebsd: udp-spoof-freebsd

udp-spoof-freebsd: $(wildcard cmd/udp-spoof/* pkg/spoof/*)
	GOOS=freebsd go build -o udp-spoof-freebsd ./cmd/udp-spoof/...

# run everything but `lint` because that runs via it's own workflow
.build-tests: vet test-fmt

.PHONY: vet
vet: ## Run `go vet` on the code
	@echo checking code is vetted...
	for x in $(shell go list ./...); do echo $$x ; go vet $$x ; done

.PHONY: fmt
fmt: ## Format Go code
	@gofmt -s -w */*.go

.PHONY: test-fmt
test-fmt: fmt ## Test to make sure code if formatted correctly
	@if test `git diff spoofrawip | wc -l` -gt 0; then \
	    echo "Code changes detected when running 'go fmt':" ; \
	    git diff -Xfiles ; \
	    exit -1 ; \
	fi

.PHONY: test-tidy
test-tidy:  ## Test to make sure go.mod is tidy
	@go mod tidy
	@if test `git diff go.mod | wc -l` -gt 0; then \
	    echo "Need to run 'go mod tidy' to clean up go.mod" ; \
	    exit -1 ; \
	fi

lint:  ## Run golangci-lint
	golangci-lint run
