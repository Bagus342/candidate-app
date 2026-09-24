FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/candidate-app .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates wget

WORKDIR /app
COPY --from=builder /out/candidate-app ./candidate-app
COPY web ./web

ENV HTTP_ADDR=:8080
EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=5 CMD wget -q --spider http://127.0.0.1:8080/ || exit 1

USER nobody
ENTRYPOINT ["/app/candidate-app"]
