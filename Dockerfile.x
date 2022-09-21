FROM ubuntu:latest
ARG TARGETOS
ARG TARGETARCH

WORKDIR /didem

COPY "build/didem_${TARGETOS}_${TARGETARCH}/didem" .

RUN chmod +x /didem/didem

CMD [ "/didem/didem" ]