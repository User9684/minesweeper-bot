# build the app
FROM golang:alpine AS builder

WORKDIR $GOPATH/src/main

COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /go/bin/app -v ./app

# start the app
FROM alpine:latest

RUN addgroup -S app && adduser -S app -G app
COPY --from=builder --chown=app /go/bin/app /app
USER app

ENTRYPOINT ["/app"]