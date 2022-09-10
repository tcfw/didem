FROM ubuntu:latest
ARG TARGETOS
ARG TARGETARCH

WORKDIR /didem

COPY "build/didem_${TARGETOS}_${TARGETARCH}" ./app

CMD [ "/didem/app" ]