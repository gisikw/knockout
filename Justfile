default:
    @just --list

build:
    go build -ldflags="-X main.version=$(git rev-parse --short HEAD)" -o ko

test:
    go test ./...

install:
    go install -ldflags="-X main.version=$(git rev-parse --short HEAD)"
