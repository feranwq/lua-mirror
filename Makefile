GOPROXY=https://goproxy.cn"
GO111MODULE=on
VERSION=$(shell git describe --abbrev=0 --tags)
COMMIT=$(shell git rev-parse --short HEAD)
PREFIX="luamirror"

.PHONY: build fmt vendor check

build: fmt check vendor
	@echo "start build"
	@cd ./cmd && go build -o ../${PREFIX} .

fmt :
	@echo "format code"
	@gofmt -l -w ./

vendor :
	@echo "go mod vendor"
	@go mod vendor

check :
	@echo "check code"
	@go vet  ./...
