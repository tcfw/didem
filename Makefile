GOBIN:=go
GOFLAGS:=-ldflags="-w -s" -trimpath
BUILD_DST:=./build
BIN:=didem

.PHONY: build
build:
	@mkdir -p ${BUILD_DST}
	${GOBIN} build ${GOFLAGS} -o ${BUILD_DST}/${BIN} ./cmd/

.PHONY: compress
compress:
	upx -9 ${BUILD_DST}/*