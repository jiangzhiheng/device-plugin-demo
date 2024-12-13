FROM golang:1.23.4 AS builder

WORKDIR /app

# Copy go.mod and go.sum files and download dependencies
COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download

# Copy the entire project
COPY . .

# Build the project
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/device-plugin-demo main.go

FROM mirrors.tencent.com/tlinux/tlinux2.4-minimal:latest

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/bin/device-plugin-demo .

ENTRYPOINT ["./device-plugin-demo"]