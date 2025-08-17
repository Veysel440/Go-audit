FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -trimpath -ldflags="-s -w" -o /api ./cmd/api

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /api /api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/api"]