FROM golang:1.23@sha256:adfbe17b774398cb090ad257afd692f2b7e0e7aaa8ef0110a48f0a775e3964f4 as build

WORKDIR /go/src/github.com/kubecano/cano-collector

COPY go.mod go.sum ./
RUN go mod download

COPY main.go .
RUN go vet -v && go test -v

RUN CGO_ENABLED=0 go build -o /go/bin/cano-collector

FROM gcr.io/distroless/static-debian12@sha256:3f2b64ef97bd285e36132c684e6b2ae8f2723293d09aae046196cca64251acac

LABEL author="KubeCano Team"
LABEL contact="support@kubecano.com"
LABEL license="Apache-2.0"
LABEL org.opencontainers.image.title="cano-collector"
LABEL org.opencontainers.image.source="https://github.com/kubecano/cano-collector"

COPY --from=build /go/bin/cano-collector /
CMD ["/cano-collector"]
