SQLITE_DESTINATION_VERSION ?= v2.14.9
GOLANGCI_LINT_VERSION ?= v2.13.1
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo development)
LDFLAGS := -s -w -X cq-source-mws/plugin.Version=$(VERSION)

.PHONY: build
build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o cq-source-mws .

.PHONY: test
test:
	go test -race -count=1 ./...

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./...

.PHONY: tidy
tidy:
	go mod tidy
	git diff --exit-code -- go.mod go.sum

.PHONY: vuln
vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# The SQLite destination is MPL-2.0 and lives in the CloudQuery monorepo.
# Building it here keeps a sync free of CloudQuery Hub, which otherwise
# requires an account: `cloudquery login` locally or CLOUDQUERY_API_KEY in CI.
.PHONY: destination
destination: dist/cq-destination-sqlite

dist/cq-destination-sqlite:
	rm -rf build/cloudquery
	git clone --depth 1 --filter=blob:none --sparse \
		--branch plugins/destination/sqlite/$(SQLITE_DESTINATION_VERSION) \
		https://github.com/cloudquery/cloudquery.git build/cloudquery
	cd build/cloudquery && git sparse-checkout set plugins/destination/sqlite
	cd build/cloudquery/plugins/destination/sqlite && go build -o $(CURDIR)/$@ .

# Requires MWS credentials: MWS_TOKEN or MWS_SERVICE_ACCOUNT_AUTHORIZED_KEY_PATH.
.PHONY: sync
sync: build destination
	mkdir -p test/output
	cloudquery sync test/config.yaml --log-level warn

.PHONY: clean
clean:
	rm -rf dist build cq-source-mws test/output
