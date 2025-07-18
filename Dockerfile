FROM golang:1.24@sha256:14fd8a55e59a560704e5fc44970b301d00d344e45d6b914dda228e09f359a088 as build

WORKDIR /go/src/github.com/kubecano/cano-collector

COPY go.mod go.sum ./
RUN go mod download

COPY main.go .
COPY config/ ./config/
COPY pkg/ ./pkg/

RUN CGO_ENABLED=0 go build -o /go/bin/cano-collector

FROM gcr.io/distroless/static-debian12@sha256:b7b9a6953e7bed6baaf37329331051d7bdc1b99c885f6dbeb72d75b1baad54f9

LABEL author="KubeCano Team"
LABEL contact="support@kubecano.com"
LABEL license="Apache-2.0"
LABEL org.opencontainers.image.title="cano-collector"
LABEL org.opencontainers.image.source="https://github.com/kubecano/cano-collector"

EXPOSE 8080

COPY --from=build /go/bin/cano-collector /
CMD ["/cano-collector"]
