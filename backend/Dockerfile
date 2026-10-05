FROM golang:1.27.1-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/server /server
COPY --from=builder /out/migrate /migrate
COPY migrations /migrations

ENV MIGRATIONS_PATH=/migrations

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/server"]
