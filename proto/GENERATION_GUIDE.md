# Proto Code Generation Guide

This guide explains how to generate code from proto files for each service.

## Quick Start

### For Go Services (account-service, identity-service, fraud-service)

```bash
cd services/<service-name>
make proto-install  # Install dependencies (first time only)
make proto          # Generate code
```

### For Java Services (customer-service)

```bash
cd services/customer-service
make proto-install  # Resolve Maven dependencies
make proto          # Generate Java code (uses Maven)
```

### For Node.js Services (transaction-service)

**No code generation needed!** Node.js loads proto files at runtime.

```bash
cd services/transaction-service
make proto  # Just validates proto files exist
```

## Service-Specific Instructions

### account-service (Go)

```bash
cd services/account-service
make proto-install  # Install protoc-gen-go and protoc-gen-go-grpc
make proto          # Generates code in proto/account/
```

**Generated files:**
- `proto/account/account.pb.go` - Message types
- `proto/account/account_grpc.pb.go` - gRPC service stubs

### identity-service (Go)

```bash
cd services/identity-service
make proto-install
make proto          # Generates code in proto/identity/
```

**Generated files:**
- `proto/identity/identity.pb.go`
- `proto/identity/identity_grpc.pb.go`

### customer-service (Java)

```bash
cd services/customer-service
make proto          # Uses Maven protobuf plugin
```

**Generated files:**
- `target/generated-sources/protobuf/java/com/bankflow/customer/proto/`

**Note:** Maven automatically runs proto generation during `mvn compile`

### transaction-service (Node.js)

**No generation needed!** Proto files are loaded at runtime using `@grpc/proto-loader`.

The service uses:
- `src/clients/account.grpc.client.ts` - gRPC client implementation
- `src/clients/account.client.ts` - Client wrapper/factory

### fraud-service (Go - when implemented)

```bash
cd services/fraud-service
make proto-install
make proto          # Generates code in proto/fraud/
```

## Common Commands

### Validate All Proto Files

```bash
cd proto
make validate
```

### Clean Generated Files

```bash
cd services/<service-name>
make proto-clean
```

## Troubleshooting

### "protoc: command not found"

Install Protocol Buffers compiler:
- **macOS:** `brew install protobuf`
- **Linux:** `apt-get install protobuf-compiler`
- **Windows:** Download from https://github.com/protocolbuffers/protobuf/releases

### "protoc-gen-go: program not found"

Run `make proto-install` in the service directory to install Go proto plugins.

### Java: "protobuf-maven-plugin not found"

The plugin is configured in `pom.xml`. Run `mvn dependency:resolve` to download it.

### Node.js: "Cannot find module '../clients/account.client'"

The wrapper file `account.client.ts` should exist. If missing, check that all files in `src/clients/` are present.

## File Structure After Generation

### Go Services
```
services/account-service/
├── proto/
│   └── account/
│       ├── account.pb.go
│       └── account_grpc.pb.go
└── Makefile
```

### Java Services
```
services/customer-service/
├── target/
│   └── generated-sources/
│       └── protobuf/
│           └── java/
│               └── com/
│                   └── bankflow/
│                       └── customer/
│                           └── proto/
└── pom.xml
```

### Node.js Services
```
services/transaction-service/
├── src/
│   └── clients/
│       ├── account.grpc.client.ts  # Runtime proto loading
│       └── account.client.ts        # Client wrapper
└── Makefile
```

## Integration with Build Process

### Go Services
Add to your build script:
```bash
make proto  # Generate proto code
go build    # Build service
```

### Java Services
Maven automatically generates proto code during `mvn compile`:
```bash
mvn clean compile  # Generates proto code and compiles
```

### Node.js Services
No special build step needed - proto files are loaded at runtime.

## Best Practices

1. **Always commit generated files** for Go/Java services (makes builds reproducible)
2. **Don't commit generated files** for Node.js (they're not needed)
3. **Run `make proto`** before building services that use gRPC
4. **Update proto files** in the `proto/` directory, not in service directories
5. **Regenerate code** after updating proto files

