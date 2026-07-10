FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -o /prest-mcp ./cmd/prest-mcp

FROM alpine:3.21

LABEL io.modelcontextprotocol.server.name="io.github.prest/prest"

COPY --from=builder /prest-mcp /usr/local/bin/prest-mcp

ENTRYPOINT ["prest-mcp"]
