# Build stage
FROM maven:3.9-eclipse-temurin-17 AS builder

WORKDIR /app

# Copy pom.xml and download dependencies
COPY pom.xml .
RUN mvn dependency:go-offline -B

# Copy source code
COPY src ./src

# Build the application
RUN mvn clean package -DskipTests -B

# Run stage
FROM eclipse-temurin:17-jre-jammy

RUN apt-get update && apt-get install -y --no-install-recommends \
  wget \
  ca-certificates \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the JAR file from builder
COPY --from=builder /app/target/customer-service-*.jar app.jar

# Expose port
EXPOSE 8002

# Run as non-root user for Debian/Ubuntu
RUN groupadd -g 1000 worker && useradd -u 1000 -g worker -m -s /bin/bash worker
RUN chown -R worker:worker /app
USER worker

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8002/health || exit 1

# Run the application
ENTRYPOINT ["java", "-jar", "app.jar"]

