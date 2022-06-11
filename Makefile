GOBIN:=go
DOCKERBIN:=docker
GOFLAGS:=-ldflags="-w -s" -trimpath
BUILD_DST:=./build
BIN:=didem
IMG:=didem
IMGVER:=latest
IMGTAG:=$(IMG):$(IMGVER)

.PHONY: build
build:
	@mkdir -p ${BUILD_DST}
	${GOBIN} build ${GOFLAGS} -o ${BUILD_DST}/${BIN} ./cmd/

.PHONY: compress
compress:
	upx -9 ${BUILD_DST}/*

.PHONY: docker
docker:
	$(DOCKERBIN) build -t $(IMGTAG) .