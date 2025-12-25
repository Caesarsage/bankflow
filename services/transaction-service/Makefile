.PHONY: proto proto-clean proto-install help

# Proto file paths (relative to project root)
PROTO_ROOT = ../../proto
ACCOUNT_PROTO = $(PROTO_ROOT)/account/account.proto
TRANSACTION_PROTO = $(PROTO_ROOT)/transaction/transaction.proto

# Node.js doesn't require code generation - proto files are loaded at runtime
# This Makefile is for documentation and future use if needed

help:
	@echo "Available targets:"
	@echo "  proto-install  - Install proto dependencies (if using code generation)"
	@echo "  proto         - Validate proto files (Node.js uses runtime loading)"
	@echo "  proto-clean   - Clean generated files (if any)"

proto-install:
	@echo "Node.js services use runtime proto loading via @grpc/proto-loader"
	@echo "No code generation needed. Proto files are loaded from:"
	@echo "  - $(ACCOUNT_PROTO)"
	@echo "  - $(TRANSACTION_PROTO)"
	@npm install @grpc/grpc-js @grpc/proto-loader

proto:
	@echo "Validating proto files..."
	@if [ ! -f "$(ACCOUNT_PROTO)" ]; then \
		echo "Error: $(ACCOUNT_PROTO) not found"; \
		exit 1; \
	fi
	@if [ ! -f "$(TRANSACTION_PROTO)" ]; then \
		echo "Error: $(TRANSACTION_PROTO) not found"; \
		exit 1; \
	fi
	@echo "Proto files validated successfully"
	@echo "Note: Node.js services load proto files at runtime, no code generation needed"

proto-clean:
	@echo "No generated files to clean (Node.js uses runtime loading)"

