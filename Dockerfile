FROM golang:1.23@sha256:927112936d6b496ed95f55f362cc09da6e3e624ef868814c56d55bd7323e0959 as build

WORKDIR /go/src/github.com/kubecano/cano-collector

COPY go.* ./
RUN go mod download

COPY . .
RUN go vet -v
RUN go test -v

RUN CGO_ENABLED=0 go build -o /go/bin/cano-collector

FROM gcr.io/distroless/static-debian12@sha256:3f2b64ef97bd285e36132c684e6b2ae8f2723293d09aae046196cca64251acac

LABEL org.opencontainers.image.source="https://github.com/kubecano/cano-collector"

COPY --from=build /go/bin/cano-collector /
CMD ["/cano-collector"]
