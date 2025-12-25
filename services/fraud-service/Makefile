.PHONY: proto proto-clean proto-install help

# Proto file path (relative to project root)
PROTO_ROOT = ../../proto
FRAUD_PROTO = $(PROTO_ROOT)/fraud/fraud.proto
OUTPUT_DIR = proto/fraud

help:
	@echo "Available targets:"
	@echo "  proto-install  - Install Go proto dependencies"
	@echo "  proto         - Generate Go code from proto files"
	@echo "  proto-clean   - Clean generated proto files"

proto-install:
	@echo "Installing Go proto dependencies..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Dependencies installed successfully"

proto:
	@echo "Generating Go code from $(FRAUD_PROTO)..."
	@mkdir -p $(OUTPUT_DIR)
	@protoc --proto_path=$(PROTO_ROOT) \
		--go_out=$(OUTPUT_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(OUTPUT_DIR) \
		--go-grpc_opt=paths=source_relative \
		fraud/fraud.proto
	@echo "Go code generated successfully in $(OUTPUT_DIR)"

proto-clean:
	@echo "Cleaning generated files in $(OUTPUT_DIR)..."
	@rm -rf $(OUTPUT_DIR)/*.pb.go
	@echo "Clean complete"

