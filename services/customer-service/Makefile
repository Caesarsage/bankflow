.PHONY: proto proto-clean proto-install help

# Proto file path (relative to project root)
PROTO_ROOT = ../../proto
CUSTOMER_PROTO = $(PROTO_ROOT)/customer/customer.proto
OUTPUT_DIR = target/generated-sources/protobuf/java

help:
	@echo "Available targets:"
	@echo "  proto-install  - Install Java proto dependencies (via Maven)"
	@echo "  proto         - Generate Java code from proto files (via Maven)"
	@echo "  proto-clean   - Clean generated proto files"

proto-install:
	@echo "Java proto dependencies are managed via Maven"
	@echo "Ensure protobuf-maven-plugin is configured in pom.xml"
	@mvn dependency:resolve

proto:
	@echo "Generating Java code from $(CUSTOMER_PROTO)..."
	@mvn protobuf:compile
	@echo "Java code generated successfully in $(OUTPUT_DIR)"

proto-clean:
	@echo "Cleaning generated files..."
	@mvn clean
	@echo "Clean complete"

