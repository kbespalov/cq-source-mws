SQLITE_DESTINATION_VERSION ?= v2.14.9

.PHONY: build
build:
	go build -o cq-source-mws .

.PHONY: test
test:
	go test ./...

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
