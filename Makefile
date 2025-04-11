
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
    # We need to build the binary with the race flag enabled
    # because it will get picked up and run during e2e tests
    # and the e2e tests should error out if any kind of race is found
	go build -race -o e2e/framework/artifacts/nexus .
	NEXUS_BINARY=${PWD}/artifacts/nexus GETH_BINARY=${PWD}/artifacts/nexus-geth go test -v -timeout=30m ./e2e/...

.PHONY: run-local
run-local:
	docker-compose -f ./docker/local/docker-compose.yml up -d --build

.PHONY: stop-local
stop-local:
	docker-compose -f ./docker/local/docker-compose.yml stop

.PHONY: destroy-local
destroy-local:
	docker-compose -f ./docker/local/docker-compose.yml down -v

.PHONY: run-single stop-single clean-single rerun-single

run-single:
	docker compose -f ./e2e/tests/docker/docker-compose.single.yaml up -d

stop-single:
	docker compose -f ./e2e/tests/docker/docker-compose.single.yaml stop

clean-single: stop-single
	docker compose -f ./e2e/tests/docker/docker-compose.single.yaml down -v

rerun-single: stop-single clean-single run-single

run-multi:
	docker compose -f ./e2e/tests/docker/docker-compose.multi.yaml up -d

stop-multi:
	docker compose -f ./e2e/tests/docker/docker-compose.multi.yaml stop

clean-multi: stop-multi
	docker compose -f ./e2e/tests/docker/docker-compose.multi.yaml down -v

rerun-single: stop-multi clean-multi run-multi

set-up-prerequisites:
	echo "Run: gvm use go1.23.0 && gvm pkgset use go1.23 if needed"
	go clean -testcache
	rm -rf e2e/tests/shared && cp -r e2e/tests/template-configs e2e/tests/shared
	docker build -t nexus-dev:latest .

test-single-liveness: set-up-prerequisites
	cd e2e/tests && go test -timeout 600s -run ^TestE2ESingleLiveness github.com/apex-fusion/nexus

test-single-broadcast: set-up-prerequisites
	cd e2e/tests && go test -timeout 600s -run ^TestE2ESingleBroadcast github.com/apex-fusion/nexus

test-single: test-single-liveness test-single-broadcast

test-multi-liveness: set-up-prerequisites
	cd e2e/tests && go test -timeout 1000s -run ^TestE2EMultiLiveness github.com/apex-fusion/nexus

test-multi-broadcast: set-up-prerequisites
	cd e2e/tests && go test -timeout 1000s -run ^TestE2EMultiBroadcast github.com/apex-fusion/nexus

test-multi: test-multi-liveness test-multi-broadcast