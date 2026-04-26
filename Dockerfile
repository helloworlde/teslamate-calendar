FROM golang:1.22 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/teslamate-calendar ./cmd/teslamate-calendar

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder /out/teslamate-calendar /usr/local/bin/teslamate-calendar
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/teslamate-calendar"]
