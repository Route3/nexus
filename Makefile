
.PHONY: download-spec-tests
download-spec-tests:
	git submodule init
	git submodule update

.PHONY: bindata
bindata:
	go-bindata -pkg chain -o ./chain/chain_bindata.go ./chain/chains

.PHONY: protoc
protoc:
	protoc --go_out=. --go-grpc_out=. ./server/proto/*.proto
	protoc --go_out=. --go-grpc_out=. ./protocol/proto/*.proto
	protoc --go_out=. --go-grpc_out=. ./network/proto/*.proto
	protoc --go_out=. --go-grpc_out=. ./txpool/proto/*.proto
	protoc --go_out=. --go-grpc_out=. ./consensus/ibft/**/*.proto

.PHONY: build
build:
	$(eval LATEST_VERSION = $(shell git describe --tags --abbrev=0))
	$(eval COMMIT_HASH = $(shell git rev-parse HEAD))
	$(eval BRANCH = $(shell git rev-parse --abbrev-ref HEAD | tr -d '\040\011\012\015\n'))
	$(eval TIME = $(shell date))
	go build -o nexus -tags netgo -ldflags="\
		-s -w -linkmode external -extldflags "-static" \
    	-X 'github.com/apex-fusion/nexus/versioning.Version=$(LATEST_VERSION)' \
		-X 'github.com/apex-fusion/nexus/versioning.Commit=$(COMMIT_HASH)'\
		-X 'github.com/apex-fusion/nexus/versioning.Branch=$(BRANCH)'\
		-X 'github.com/apex-fusion/nexus/versioning.BuildTime=$(TIME)'" \
	main.go

.PHONY: lint
lint:
	golangci-lint run --config .golangci.yml

.PHONY: generate-bsd-licenses
generate-bsd-licenses:
	./generate_dependency_licenses.sh BSD-3-Clause,BSD-2-Clause > ./licenses/bsd_licenses.json

.PHONY: test
test:
	go test -timeout=20m `go list ./... | grep -v e2e`

.PHONY: test-e2e
test-e2e:
	curl -L -o nexus-geth https://github.com/Route3/nexus-geth/releases/download/v1.0.1/nexus-geth
	chmod +x nexus-geth
	mkdir -p ./e2e/framework/artifacts
	mv nexus-geth ./e2e/framework/artifacts/nexus-geth
	go test -v -timeout=30m ./e2e/...