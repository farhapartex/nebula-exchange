FROM golang:1.27.1-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/ ./cmd/server ./cmd/migrate ./cmd/seed

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/server /server
COPY --from=builder /out/migrate /migrate
COPY --from=builder /out/seed /seed
COPY migrations /migrations
COPY seeds /seeds

ENV MIGRATIONS_PATH=/migrations

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/server"]
