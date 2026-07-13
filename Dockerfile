FROM golang:1.25@sha256:698183780de28062f4ef46f82a79ec0ae69d2d22f7b160cf69f71ea8d98bf25d as build

WORKDIR /go/src/github.com/kubecano/cano-collector

COPY go.mod go.sum ./
RUN go mod download

COPY main.go .
COPY config/ ./config/
COPY pkg/ ./pkg/

RUN CGO_ENABLED=0 go build -o /go/bin/cano-collector

FROM gcr.io/distroless/static-debian12@sha256:22fd79fd75eab2372585b44517f8a094349938919dc613aafc37e4bdc9967c82

LABEL author="KubeCano Team"
LABEL contact="support@kubecano.com"
LABEL license="Apache-2.0"
LABEL org.opencontainers.image.title="cano-collector"
LABEL org.opencontainers.image.source="https://github.com/kubecano/cano-collector"

EXPOSE 8080

COPY --from=build /go/bin/cano-collector /
CMD ["/cano-collector"]
