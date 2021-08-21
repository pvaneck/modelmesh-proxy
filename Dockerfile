# Build
FROM golang:1.17-alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR /app
COPY . .
RUN go get -d -v
# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o /go/bin/server

# Runtime
FROM scratch

ARG USER=2000

COPY --from=builder /go/bin/server /go/bin/server
ENTRYPOINT ["/go/bin/server"]
