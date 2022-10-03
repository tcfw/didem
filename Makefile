GOBIN?=go
DOCKERBIN?=docker
VERSION?=$$(git rev-parse --short HEAD)
GOFLAGS?=-ldflags="-w -s -X 'github.com/tcfw/didem/internal/cli.GitVersion=${VERSION}'" -trimpath
BUILD_DST?=./build
BIN?=didem

IMGREPO?=ghcr.io/tcfw/
IMG?=didem
IMGVER?=latest
IMGTAG?=$(IMGREPO)$(IMG):$(IMGVER)

.PHONY: build
build:
	@mkdir -p ${BUILD_DST}
	${GOBIN} build ${GOFLAGS} -o ${BUILD_DST}/${BIN} ./cmd/

.PHONY: install
install:
	${GOBIN} install ${GOFLAGS} .

.PHONY: compress
compress:
	upx -9 ${BUILD_DST}/*

.PHONY: docker
docker:
	$(DOCKERBIN) build -t $(IMGTAG) .

.PHONY: docker-push
docker-push:
	$(DOCKERBIN) push $(IMGTAG)

.PHONY: gen-api
gen-api:
	@./scripts/genproto.sh

.PHONY: test
test:
	go test -v -cover -covermode atomic ./...

.PHONY: race
race:
	go test -race ./...