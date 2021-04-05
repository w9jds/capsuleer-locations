FROM golang:latest as builder

LABEL version="v2.0"
LABEL maintainer="Jeremy Shore <w9jds@github.com>"

WORKDIR /go/src/locations

COPY . /go/src/locations

RUN go get -d ./...
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-w -s' -a -installsuffix cgo -o locations
RUN curl -o ca-certificates.crt https://raw.githubusercontent.com/bagder/ca-bundle/master/ca-bundle.crt

FROM scratch

WORKDIR /go/src/locations

COPY --from=builder /go/src/locations/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/locations/config /go/src/locations/config
COPY --from=builder /go/src/locations/locations /go/src/locations

CMD ["./locations"]