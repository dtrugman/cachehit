#!/bin/bash

DIST_DIR="dist"

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;36m'
NC='\033[0m' # No Color

print_step() {
    echo -e "${BLUE}[-]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_done() {
    print_success "Done"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

help() {
    cat << EOF
Usage: $0 [command]

Commands:
    help            Display this help screen
    test            Run tests
    test-coverage   Run tests with coverage
    test-race       Run tests with race detector
    bench           Run benchmarks
    lint            Run linter
    fmt             Format code
    vet             Run go vet
    verify          Run all verification steps (fmt, vet, lint, test)
    clean           Clean build artifacts
    install-tools   Install development tools
    deps            Download dependencies
    tidy-deps       Tidy dependencies
    upgrade-deps    Upgrade dependencies
    ci              Run CI pipeline locally

EOF
}

build() {
    print_step "Building examples..."

    go build -o "$DIST_DIR/swr" "./example/swr"
    go build -o "$DIST_DIR/layered" "./example/layered"

    print_done
}

test() {
    print_step "Running tests..."

    go test "$@" ./...
}

test_coverage() {
    print_step "Running tests with coverage..."

    go test -v -coverprofile=coverage.out -covermode=atomic ./... || return 1
    go tool cover -html=coverage.out -o coverage.html || return 1

    return 0
}

test_race() {
    print_step "Running tests with race detector..."

    go test -race -short ./...
}

bench() {
    print_step "Running benchmarks..."

    go test -bench=. -benchmem ./...
}

lint() {
    print_step "Running linter..."

    if !command -v golangci-lint &> /dev/null; then
        print_error "Missing tool. Run 'install-tools' first"
    fi

    golangci-lint run ./...

    print_done
}

fmt() {
    print_step "Formatting code..."

    go fmt ./...
    gofmt -s -w .

    print_done
}

vet() {
    print_step "Vetting code..."

    go vet ./...

    print_done
}

verify() {
    print_step "Running all verification steps..."

    fmt || return 1
    vet || return 1
    lint || return 1
    test || return 1

    print_done
}

clean() {
    print_step "Cleaning..."

    rm -f coverage.out coverage.html
    go clean

    print_done
}

install_deps() {
    print_step "Downloading dependencies..."

    go mod download

    print_done
}

tidy_deps() {
    print_step "Tidying dependencies..."

    go mod tidy

    print_done
}

upgrade_deps() {
    print_step "Upgrading dependencies..."

    go get -u ./...
    go mod tidy

    print_done
}

main() {
    declare -r cmd="${1:-help}"
    case "$cmd" in
        help)
            help
            ;;
        build)
            build
            ;;
        test)
            test "${@:2}"
            ;;
        test-coverage)
            test_coverage
            ;;
        test-race)
            test_race
            ;;
        bench)
            bench
            ;;
        lint)
            lint
            ;;
        fmt)
            fmt
            ;;
        vet)
            vet
            ;;
        verify)
            verify
            ;;
        clean)
            clean
            ;;
        install-tools)
            install_tools
            ;;
        install-deps)
            install_deps
            ;;
        tidy-deps)
            tidy_deps
            ;;
        upgrade-deps)
            upgrade_deps
            ;;
        *)
            echo "Unknown command: $1"
            echo ""
            help
            exit 1
            ;;
    esac

    return $?
}

main "$@"
