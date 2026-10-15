FROM golang:1.25.2-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/cmd/levelup/exe /app/cmd/levelup

FROM alpine:3.24
WORKDIR /app
COPY --from=builder /app/cmd/levelup/exe /app
CMD ["/app/exe"]