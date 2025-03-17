FROM golang:1.24@sha256:fa145a3c13f145356057e00ed6f66fbd9bf017798c9d7b2b8e956651fe4f52da as build

WORKDIR /go/src/github.com/kubecano/cano-collector

COPY go.mod go.sum ./
RUN go mod download

COPY main.go .
COPY config/ ./config/
COPY pkg/ ./pkg/

RUN CGO_ENABLED=0 go build -o /go/bin/cano-collector

FROM gcr.io/distroless/static-debian12@sha256:3f2b64ef97bd285e36132c684e6b2ae8f2723293d09aae046196cca64251acac

LABEL author="KubeCano Team"
LABEL contact="support@kubecano.com"
LABEL license="Apache-2.0"
LABEL org.opencontainers.image.title="cano-collector"
LABEL org.opencontainers.image.source="https://github.com/kubecano/cano-collector"

EXPOSE 8080

COPY --from=build /go/bin/cano-collector /
CMD ["/cano-collector"]
