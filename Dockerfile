FROM golang:alpine@sha256:3ad57304ad93bbec8548a0437ad9e06a455660655d9af011d58b993f6f615648 AS builder

WORKDIR /app
COPY . .
RUN go build -o /app/server ./cmd/server

FROM alpine:latest@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b

RUN apk add --no-cache ca-certificates \
 && adduser -D -u 1001 appuser

WORKDIR /

COPY --from=builder /app/server /server
COPY --from=builder /app/web /web
COPY --from=builder /app/internal/dictionary /internal/dictionary

EXPOSE ${PORT:-3102}
USER appuser

CMD ["/server"]
