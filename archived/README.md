# Archived Functions

This directory contains archived KRM functions that are no longer actively maintained but are preserved for reference and compatibility.

## Building Functions

Note: We use `docker buildx` to build images. Please ensure you have it installed.

### Build All Functions

```shell
# Build all Go functions
make build-go

# Build all TypeScript functions  
make build-ts
```

### Build Individual Functions

```shell
# Build a specific function using the unified script
./archived-function-release.sh load go export-terraform latest

# Build a TypeScript function
./archived-function-release.sh load ts sops latest
```

## Testing

### Unit Tests

To run all unit tests

```shell
make unit-test
```

To run tests for specific language:

```shell
# Run Go unit tests
make unit-test-go

# Run TypeScript unit tests
make unit-test-ts
```

### E2E Tests

To test archived function examples:

```shell
make e2e-test
```

To test a specific archived example:

```shell
cd ../tests/e2etest
go test -v ./... -run TestArchivedE2E/../../archived/examples/$EXAMPLE_NAME
```

To run all tests (unit + examples):

```shell
make test
```

## Directory Structure

```
archived/
├── functions/
│   ├── go/           # Go-based archived functions
│   └── ts/           # TypeScript-based archived functions
├── examples/         # Example configurations for archived functions
├── build/            # Docker build configurations
│   └── docker/
│       ├── go/       # Go Dockerfile and defaults
│       └── ts/       # TypeScript Dockerfile and defaults
├── archived-function-release.sh  # Unified build script
└── docker-archived.sh           # Docker build helper
```

## Notes

- Archived functions use specialized build scripts due to different directory structures
- TypeScript functions may require custom Dockerfiles (e.g., `sops.Dockerfile`)
- Some functions have outdated dependencies and use `skipLibCheck` in TypeScript compilation
- Functions are published to `ghcr.io/kptdev/krm-functions-catalog/archived`