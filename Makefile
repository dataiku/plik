SHELL = bash

BUILD_INFO = $(shell server/gen_build_info.sh base64)

# External ldflags: default to -static for static builds on linux.
# macOS (Darwin) does not support full static linking; avoid -static there.
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
	EXTLDFLAGS :=
else
	EXTLDFLAGS := -static
endif

ifeq ($(strip $(EXTLDFLAGS)),)
	BUILD_FLAG = -ldflags="-X github.com/root-gg/plik/server/common.buildInfoString=$(BUILD_INFO) -w -s"
else
	BUILD_FLAG = -ldflags="-X github.com/root-gg/plik/server/common.buildInfoString=$(BUILD_INFO) -w -s -extldflags=$(EXTLDFLAGS)"
endif

BUILD_TAGS = -tags osusergo,netgo,sqlite_omit_load_extension

GO_BUILD = go build $(BUILD_FLAG) $(BUILD_TAGS)

COVER_FILE = /tmp/plik.coverprofile
GO_TEST = GORACE="halt_on_error=1" go test $(BUILD_FLAG) $(BUILD_TAGS) -race -cover -coverprofile=$(COVER_FILE) -p 1

ifdef ENABLE_RACE_DETECTOR
	GO_BUILD := GORACE="halt_on_error=1" $(GO_BUILD) -race
endif

all: clean clean-frontend frontend clients server

###
# Build frontend ressources
###
frontend:
	@cd webapp && npm ci && npm run build

###
# Build plik server for the current architecture
###
server:
	@server/gen_build_info.sh info
	@echo "Building Plik server"
	@cd server && $(GO_BUILD) -o plikd

###
# Build plik client for the current architecture
###
client:
	@server/gen_build_info.sh info
	@echo "Building Plik client"
	@cd client && $(GO_BUILD) -o plik ./

###
# Build clients for all architectures
###
clients:
	@releaser/build_clients.sh

###
# Display build info
###
build-info:
	@server/gen_build_info.sh info

###
# Display version
###
version:
	@server/gen_build_info.sh version

###
# Run linters
###
lint:
	@FAIL=0 ;echo -n " - go fmt :" ; OUT=`gofmt -l client server plik` ; \
	if [[ -z "$$OUT" ]]; then echo " OK" ; else echo " FAIL"; echo "$$OUT"; FAIL=1 ; fi ;\
	echo -n " - go vet :" ; OUT=`go vet ./... 2>&1` ; \
	if [[ -z "$$OUT" ]]; then echo " OK" ; else echo " FAIL"; echo "$$OUT"; FAIL=1 ; fi ;\
	echo -n " - go fix :" ; OUT=`go fix ./... 2>&1` ; \
	if [[ -z "$$OUT" ]]; then echo " OK" ; else echo " FAIL"; echo "$$OUT"; FAIL=1 ; fi ;\
	test $$FAIL -eq 0

###
# Run vulnerability check (requires: go install golang.org/x/vuln/cmd/govulncheck@latest)
###
vuln:
	@echo "Running govulncheck..."
	@govulncheck ./... || true

###
# Run fmt
###
fmt:
	@gofmt -w -s client server plik

###
# Run go fix
###
gofix:
	@go fix -v ./...

###
# Run tests
###
test:
	@if curl -s 127.0.0.1:8080 > /dev/null ; then echo "Plik server probably already running" ; exit 1 ; fi
	@$(GO_TEST) ./... 2>&1 | grep -v "no test files"; test $${PIPESTATUS[0]} -eq 0
	@echo "cli client integration tests :" && cd client && ./test.sh

###
# Open last cover profile in web browser
###
cover:
	@if [[ ! -f $(COVER_FILE) ]]; then echo "Please run \"make test\" first to generate a cover profile" ; exit 1; fi
	@go tool cover -html=$(COVER_FILE)
	@echo "Check your web browser to see the cover profile"

###
# Run integration tests for all available backends
###
test-backends:
	@testing/test_backends.sh

###
# Run integration tests for a single backend
# Usage: make test-backend mariadb
###
ifeq (test-backend,$(firstword $(MAKECMDGOALS)))
  BACKEND_ARG := $(wordlist 2,2,$(MAKECMDGOALS))
  $(eval $(BACKEND_ARG):;@:)
endif

test-backend:
	@testing/test_backends.sh $(BACKEND_ARG)

###
# Build documentation
###
docs:
	@cd docs && npm ci && bash inject_version.sh && bash copy_architecture.sh && npm run build

###
# Build a docker image locally
###
docker:
	@docker buildx build --progress=plain --load -t rootgg/plik:dev .

###
# Create release archives
###
release:
	@releaser/release.sh

###
# Create release archives, build a multiarch Docker image and push to Docker Hub
###
release-and-push-to-docker-hub:
	@PUSH_TO_DOCKER_HUB=true releaser/release.sh

###
# Remove server build files
###
clean:
	@rm -rf server/plikd
	@rm -rf client/plik
	@rm -rf clients
	@rm -rf servers
	@rm -rf release
	@rm -rf releases

###
# Remove frontend build files
###
clean-frontend:
	@rm -rf webapp/dist

###
# Remove all build files and node modules
###
clean-all: clean clean-frontend
	@rm -rf webapp/node_modules

###
# Since the client/server/version directories are not generated
# by make, we must declare these targets as phony to avoid :
# "make: `client' is up to date" cases at compile time
###
.PHONY: client clients server release docs test-backend
