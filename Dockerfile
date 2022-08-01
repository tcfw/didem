#builder
FROM docker.io/golang:1.17-bullseye as build

WORKDIR /go/src/app
ADD go.mod /go/src/app
ADD go.sum /go/src/app

RUN go mod download -x

ADD . /go/src/app

RUN make build

#
FROM gcr.io/distroless/base-debian11:debug

COPY --from=build /go/src/app/build/didem /

RUN ["/busybox/mkdir", "-p", "/root/.didem/"]
RUN ["/busybox/touch", "/root/.didem/identities.yaml"]
ADD debug/cfg.yaml /root/.didem/didem.yaml

ENTRYPOINT ["/didem"]