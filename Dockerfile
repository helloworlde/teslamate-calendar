ARG GO_VERSION=1.22

FROM golang:${GO_VERSION}-bookworm AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" \
    -o /out/teslamate-calendar ./cmd/teslamate-calendar

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=builder /out/teslamate-calendar /teslamate-calendar
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/teslamate-calendar"]
