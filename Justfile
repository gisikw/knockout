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
    fort ratched systemd '{"action": "restart", "unit": "knockout"}'

agent-session-complete:
    git push
    just restart

# Symlink this project's pipeline and prompts into another project.
# Usage: just link-pipeline nerve
link-pipeline project:
    #!/usr/bin/env bash
    set -euo pipefail
    target="{{ project }}"
    # Resolve bare names as ~/Projects/<name>
    if [[ "$target" != /* && "$target" != ~* && "$target" != .* ]]; then
        target="$HOME/Projects/$target"
    fi
    if [ ! -d "$target" ]; then
        echo "error: $target does not exist" >&2
        exit 1
    fi
    src="{{ justfile_directory() }}/.ko"
    # Ensure target .ko dir exists
    mkdir -p "$target/.ko"
    # pipeline.yml
    if [ -e "$target/.ko/pipeline.yml" ] && [ ! -L "$target/.ko/pipeline.yml" ]; then
        echo "backing up $target/.ko/pipeline.yml → pipeline.yml.bak"
        mv "$target/.ko/pipeline.yml" "$target/.ko/pipeline.yml.bak"
    fi
    ln -sf "$src/pipeline.yml" "$target/.ko/pipeline.yml"
    echo "linked pipeline.yml"
    # prompts/
    if [ -d "$target/.ko/prompts" ] && [ ! -L "$target/.ko/prompts" ]; then
        echo "backing up $target/.ko/prompts → prompts.bak"
        mv "$target/.ko/prompts" "$target/.ko/prompts.bak"
    fi
    ln -sf "$src/prompts" "$target/.ko/prompts"
    echo "linked prompts/"
    echo "done: $target/.ko → $src"
