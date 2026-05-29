FROM golang:1.26-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o gpu-exporter ./cmd/gpu-exporter/

FROM gcr.io/distroless/cc-debian12
COPY --from=builder /app/gpu-exporter /gpu-exporter
EXPOSE 9835
ENTRYPOINT ["/gpu-exporter"]
# run: docker run --rm --gpus all -p 9835:9835 --pid=host gpu-exporter
