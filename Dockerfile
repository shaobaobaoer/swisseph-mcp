FROM golang:1.21-bookworm AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build

FROM debian:bookworm-slim
COPY --from=builder /app/bin/swisseph-mcp /usr/local/bin/
COPY --from=builder /app/third_party/swisseph/ephe /usr/local/share/swisseph/ephe

ENV SWISSEPH_EPHE_PATH=/usr/local/share/swisseph/ephe
ENTRYPOINT ["swisseph-mcp"]
