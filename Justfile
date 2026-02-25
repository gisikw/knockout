default:
    @just --list

build:
    go build -ldflags="-X main.version=$(git rev-parse --short HEAD)" -o ko

test:
    go test ./...

install:
    go build -ldflags="-X main.version=$(git rev-parse --short HEAD)" -o $(go env GOPATH)/bin/ko

restart:
    just install
    fort ratched systemd restart knockout
