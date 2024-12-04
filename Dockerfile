# Build stage
FROM golang:alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY *.go .

# Build the Go app
RUN go build -ldflags "-s -w" -o main .

# Final stage
FROM alpine:latest

# 设置时区

ENV TZ=Asia/Shanghai

RUN sed -i 's#https\?://dl-cdn.alpinelinux.org/alpine#http://mirrors.tuna.tsinghua.edu.cn/alpine#g' /etc/apk/repositories

RUN apk add --no-cache tzdata && \
    ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && \
    echo $TZ > /etc/timezone

# Set the Current Working Directory inside the container
WORKDIR /

COPY index.html /index.html

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Expose port 8080 to the outside world
EXPOSE 5340

# Command to run the executable
CMD ["./main", "-c", "config.json"]