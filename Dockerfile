FROM golang:1.23 as build

WORKDIR /go/src/github.com/kubecano/cano-collector

COPY go.* ./
RUN go mod download

COPY . .
RUN go vet -v
RUN go test -v

RUN CGO_ENABLED=0 go build -o /go/bin/cano-collector

FROM gcr.io/distroless/static-debian12

LABEL org.opencontainers.image.source="https://github.com/kubecano/cano-collector"

COPY --from=build /go/bin/cano-collector /
CMD ["/cano-collector"]
