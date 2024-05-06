INSTALL_DIR := env('INSTALL_DIR', '/usr/local/bin')

# build gatecheck binary
build:
    mkdir -p bin
    go build -ldflags="-X 'main.cliVersion=$(git describe --tags || echo "")' -X 'main.gitCommit=$(git rev-parse HEAD)' -X 'main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)' -X 'main.gitDescription=$(git log -1 --pretty=%B)'" -o ./bin/deploykit ./cmd/deploykit

build-image image-name:
    docker build --file Dockerfile --tag {{ image-name }} --build-arg VERSION=$(git describe --tags || echo "") --build-arg GIT_COMMIT=$(git rev-parse HEAD) --build-arg BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) --build-arg "GIT_DESCRIPTION=$(git log -1 --pretty=%B)" .

# build and install binary
install: build
    cp ./bin/deploykit {{ INSTALL_DIR }}/deploykit

# unit testing with coverage
test:
    go test -cover ./...

# golangci-lint view only
lint:
    golangci-lint run --fast

# golangci-lint fix linting errors and format if possible
fix:
    golangci-lint run --fast --fix

release-snapshot:
    goreleaser release --snapshot --clean

release:
    goreleaser release --clean
