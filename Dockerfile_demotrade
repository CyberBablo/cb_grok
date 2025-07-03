FROM golang:1.24-alpine3.22 as builder

WORKDIR /build

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -a -installsuffix cgo -o /main cmd/trade/main.go

FROM alpine:3
COPY --from=builder main /app/main

ENTRYPOINT ["/app/main", "--trading_mode=demo"]
