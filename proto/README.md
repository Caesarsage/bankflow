# Protocol Buffers (gRPC) Implementation Guide

## Overview

This directory contains Protocol Buffer (`.proto`) files that define gRPC services for inter-service communication in the BankFlow microservices architecture.

## How Proto Files Work with Multiple Languages

### Key Concept: Language-Specific Options

**Proto files can include options for MULTIPLE languages simultaneously.** Each language's protoc compiler:
- **Uses** its own language-specific options
- **Ignores** options for other languages

This means you can have ONE proto file that works for Go, Java, Node.js, Python, etc.

### Example

```protobuf
syntax = "proto3";

package account;

// Go compiler uses this, Java/Node.js ignore it
option go_package = "github.com/Caesarsage/bankflow/account-service/proto/account";

// Java compiler uses these, Go/Node.js ignore them
option java_package = "com.bankflow.account.proto";
option java_outer_classname = "AccountProto";
option java_multiple_files = true;
```

### Why This Works

1. **Go's protoc compiler** reads `go_package` and generates Go code in that package
2. **Java's protoc compiler** reads `java_package` and `java_outer_classname` and generates Java classes
3. **Node.js's protoc compiler** doesn't need special options - it uses the `package` name

Each compiler only processes options it understands and ignores the rest.

## Service-to-Language Mapping

| Service | Language | Proto File | Uses Options |
|---------|----------|------------|--------------|
| **identity-service** | Go | `identity/identity.proto` | `go_package` |
| **account-service** | Go | `account/account.proto` | `go_package` |
| **customer-service** | Java | `customer/customer.proto` | `java_package`, `java_outer_classname` |
| **transaction-service** | Node.js/TypeScript | `transaction/transaction.proto` | None (uses `package` name) |
| **fraud-service** | TBD | `fraud/fraud.proto` | All options included |

## Generating Code from Proto Files

### For Go Services

```bash
# Install protoc-gen-go and protoc-gen-go-grpc
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate Go code
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/account/account.proto
```

### For Java Services

```bash
# Add to pom.xml (Maven)
<plugin>
  <groupId>org.xolstice.maven.plugins</groupId>
  <artifactId>protobuf-maven-plugin</artifactId>
  <version>0.6.1</version>
  <configuration>
    <protoSourceRoot>${project.basedir}/../../proto</protoSourceRoot>
  </configuration>
  <executions>
    <execution>
      <goals>
        <goal>compile</goal>
      </goals>
    </execution>
  </executions>
</plugin>

# Generate Java code
mvn protobuf:compile
```

### For Node.js/TypeScript Services

```bash
# Install dependencies
npm install @grpc/grpc-js @grpc/proto-loader

# Use proto-loader at runtime (no code generation needed)
import * as protoLoader from '@grpc/proto-loader';
const packageDefinition = protoLoader.loadSync('proto/account/account.proto');
```

## Best Practices

1. **Include all language options** - Even if you only use one language now, include options for others to make the proto file future-proof
2. **Use consistent naming** - Follow Java package naming conventions for `java_package`
3. **Document the service** - Add comments explaining what each service does
4. **Version your proto files** - Consider adding version numbers for breaking changes

## Common Options Explained

### Go Options
- `go_package`: Full Go import path where generated code will be placed

### Java Options
- `java_package`: Java package name for generated classes
- `java_outer_classname`: Name of the outer wrapper class
- `java_multiple_files`: Generate separate file per message (recommended)

### Node.js
- No special options needed - uses the `package` declaration

## Troubleshooting

### "go_package not found" in Java
- **This is normal!** Java compiler ignores `go_package`. Use `java_package` instead.

### "java_package not found" in Go
- **This is normal!** Go compiler ignores `java_package`. Use `go_package` instead.

### Node.js can't find types
- Node.js uses dynamic loading - make sure the proto file path is correct
- Use `keepCase: true` in protoLoader options to preserve field names

## File Structure

```
proto/
├── README.md (this file)
├── account/
│   └── account.proto
├── customer/
│   └── customer.proto
├── fraud/
│   └── fraud.proto
├── identity/
│   └── identity.proto
└── transaction/
    └── transaction.proto
```

## References

- [Protocol Buffers Language Guide](https://developers.google.com/protocol-buffers/docs/proto3)
- [gRPC Documentation](https://grpc.io/docs/)
- [Go gRPC Quick Start](https://grpc.io/docs/languages/go/quickstart/)
- [Java gRPC Quick Start](https://grpc.io/docs/languages/java/quickstart/)
- [Node.js gRPC Quick Start](https://grpc.io/docs/languages/node/quickstart/)

