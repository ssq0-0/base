FROM golang:1.22 AS builder

WORKDIR /base

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o base ./app/main.go

FROM alpine:latest

WORKDIR /base
RUN mkdir -p /app/config /base/modules/abis
RUN mkdir -p /app/config

COPY --from=builder /base/base /base/base
COPY modules/abis /base/modules/abis
COPY config/config.json /base/config/config.json

RUN chmod +x /base/base

CMD ["/base/base"]
