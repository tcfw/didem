#builder
FROM docker.io/golang:1.17-bullseye as build

WORKDIR /go/src/app
ADD . /go/src/app

RUN go get -d -v ./...

RUN make build

#
FROM gcr.io/distroless/base-debian11

COPY --from=build /go/src/app/build/didem /

CMD ["/didem"]