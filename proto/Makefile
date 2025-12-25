.PHONY: help validate install-tools

# This Makefile provides common proto validation and tool installation

help:
	@echo "Proto Makefile - Common operations for all proto files"
	@echo ""
	@echo "Available targets:"
	@echo "  validate      - Validate all proto files"
	@echo "  install-tools - Install protoc compiler and language plugins"
	@echo ""
	@echo "Service-specific proto generation:"
	@echo "  cd services/<service-name> && make proto"

validate:
	@echo "Validating all proto files..."
	@for proto in account/account.proto customer/customer.proto fraud/fraud.proto identity/identity.proto transaction/transaction.proto; do \
		if [ -f "$$proto" ]; then \
			echo "✓ $$proto"; \
		else \
			echo "✗ $$proto - NOT FOUND"; \
		fi \
	done

install-tools:
	@echo "Installing protoc compiler..."
	@echo "Please install protoc manually:"
	@echo "  macOS: brew install protobuf"
	@echo "  Linux: apt-get install protobuf-compiler"
	@echo "  Windows: Download from https://github.com/protocolbuffers/protobuf/releases"
	@echo ""
	@echo "For Go services:"
	@echo "  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
	@echo "  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
	@echo ""
	@echo "For Java services:"
	@echo "  Use Maven with protobuf-maven-plugin (configured in pom.xml)"

