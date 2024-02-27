
# Change these variables as necessary.
MAIN_PACKAGE_PATH := ./examples/simple
BINARY_NAME := gt-telemetry


# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

.PHONY: no-dirty
no-dirty:
	git diff --exit-code


# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

## audit: run quality control checks
.PHONY: audit
audit:
	go mod verify
	go vet ./ ./internal/utils   # ignore Kaitai Struct files as they trip some rules
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	go test -race -buildvcs -vet=off ./...

## lint: run linters
.PHONY: lint
lint:
	golangci-lint run


# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## test: run all tests
.PHONY: test
test:
	go test -v -race -buildvcs ./...

## test/cover: run all tests and display coverage
.PHONY: test/cover
test/cover:
	go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

## kaitai: compile the GT telemetry package from the Kaitai Struct
.PHONY: build/kaitai
build/kaitai:
ifeq (, $(shell which kaitai-struct-compilerx))
	$(error "kaitai-struct-compiler command not found, see https://kaitai.io/#download for installation instructions.")
endif
	@kaitai-struct-compiler --target go --go-package gttelemetry --outdir internal internal/kaitai/gran_turismo_telemetry.ksy

## build: build the application
.PHONY: build
build:
	@go build -o examples/bin/${BINARY_NAME} ${MAIN_PACKAGE_PATH}

.PHONY: build-darwin
build-darwin:
	@GOOS=darwin GOARCH=arm64 go build -o examples/bin/${BINARY_NAME}-darwin-arm64 ${MAIN_PACKAGE_PATH}

.PHONY: build-linux
build-linux:
	@GOOS=linux GOARCH=amd64 go build -o examples/bin/${BINARY_NAME}-linux-amd64 ${MAIN_PACKAGE_PATH}

.PHONY: build-rpi
build-rpi:
	@GOOS=linux GOARCH=arm64 go build -o examples/bin/${BINARY_NAME}-rpi-arm64 ${MAIN_PACKAGE_PATH}

.PHONY: build-windows
build-windows:
	@GOOS=windows GOARCH=amd64 go build -o examples/bin/${BINARY_NAME}-amd64.exe ${MAIN_PACKAGE_PATH}

## run: run the  application
.PHONY: run
run: build
	@./examples/bin/${BINARY_NAME}

## run/live: run the application with reloading on file changes
.PHONY: run/live
run/live:
	@go run ${MAIN_PACKAGE_PATH}/main.go

## clean: clean up project and return to a pristine state
.PHONY: clean
clean:
	@go clean
	@rm -rf examples/bin