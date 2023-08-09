NAME=github.com/goto/dex
VERSION=$(shell git describe --tags --always --first-parent 2>/dev/null)
COMMIT=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date)
COVERAGE_DIR=coverage
BUILD_DIR=dist
EXE=dex

.PHONY: all format clean build test test-coverage

all: clean format test lint build 

tidy:
	@echo "Tidy up go.mod..."
	@go mod tidy -v

install:
	@echo "Installing dex to ${GOBIN}..."
	@go install
	
format:
	@echo "Running gofumpt..."
	@gofumpt -l -w .
	@echo "Running gci..."
	@gci write --skip-generated -s standard -s default -s "prefix(github.com/goto/dex)" .

lint:
	@echo "Running lint checks using golangci-lint..."
	@golangci-lint run

generate:
	@echo "Running go-generate..."
	@go generate ./...
	@echo "Cleanup old swagger output..."
	@rm -rf generated/
	@mkdir generated
	@swagger generate client -t generated -f swagger.yml
	@make format

generate-mocks:
	@mockery --srcpkg=buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/siren/v1beta1/sirenv1beta1grpc --name=SirenServiceClient
	@mockery --srcpkg=buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/shield/v1beta1/shieldv1beta1grpc --name=ShieldServiceClient
	@mockery --srcpkg=buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/optimus/core/v1beta1/corev1beta1grpc --name=JobSpecificationServiceClient

clean: tidy
	@echo "Cleaning up build directories..."
	@rm -rf ${COVERAGE_DIR} ${BUILD_DIR}

test: tidy
	@mkdir -p ${COVERAGE_DIR}
	@echo "Running unit tests..."
	@go test ./... -coverprofile=${COVERAGE_DIR}/coverage.out

test-coverage: test
	@echo "Generating coverage report..."
	@go tool cover -html=${COVERAGE_DIR}/coverage.out

build:
	@goreleaser build --snapshot --rm-dist --single-target

download:
	@go mod download

setup:
	@go install github.com/vektra/mockery/v2@v2.30.1
	@go install mvdan.cc/gofumpt@v0.5.0
