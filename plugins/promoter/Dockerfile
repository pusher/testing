# Build the manager binary
FROM golang:1.13 as builder

# Copy in the go src
WORKDIR /go/src/promoter
COPY go.mod go.mod
COPY go.sum go.sum

# Fetch dependencies before copying code (should cache unless go.sum changes)
RUN go mod download

ARG VERSION=undefined

COPY *.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o promoter -ldflags="-X main.VERSION=${VERSION}"

# Copy the controller-manager into a thin image
FROM alpine:3.10
RUN apk --no-cache add ca-certificates git
WORKDIR /bin
COPY --from=builder /go/src/promoter/promoter .
ENTRYPOINT ["/bin/promoter"]
