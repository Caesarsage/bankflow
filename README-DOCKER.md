# Docker Compose Setup for BankFlow

This Docker Compose configuration allows you to run all BankFlow microservices locally for development and testing without Kubernetes.

## Quick Start

### 1. Build and Start All Services

```bash
docker compose up --build
```

### 2. Start in Background (Detached Mode)

```bash
docker compose up -d --build
```

### 3. View Logs

```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f identity-service
docker compose logs -f customer-service
docker compose logs -f account-service
docker compose logs -f transaction-service
```

### 4. Stop All Services

```bash
docker compose down
```

### 5. Stop and Remove Volumes (Clean Slate)

```bash
docker compose down -v
```

## Services and Ports

| Service | Port | Health Check | Description |
|---------|------|--------------|-------------|
| **PostgreSQL** | 5432 | `/health` | Database for all services |
| **Redis** | 6379 | `/health` | Cache for identity service |
| **Kafka** | 9092, 9093 | `/health` | Message broker |
| **Zookeeper** | 2181 | `/health` | Kafka coordination |
| **Identity Service** | 8001 | `http://localhost:8001/health` | Authentication & authorization |
| **Customer Service** | 8002 | `http://localhost:8002/health` | Customer management & KYC |
| **Account Service** | 8004 (HTTP), 50051 (gRPC) | `http://localhost:8004/health` | Account management |
| **Transaction Service** | 8003 | `http://localhost:8003/health` | Transaction processing |

## Service URLs

- **Identity Service API**: `http://localhost:8001/api/v1`
- **Customer Service API**: `http://localhost:8002/api/v1`
- **Account Service API**: `http://localhost:8004/api/v1`
- **Account Service gRPC**: `localhost:50051`
- **Transaction Service API**: `http://localhost:8003/api/v1`

## Database Access

Connect to PostgreSQL:
```bash
docker exec -it bankflow-postgres psql -U bankflow -d identity_db
```

Available databases:
- `identity_db` - Identity service
- `customer_db` - Customer service
- `account_db` - Account service
- `transaction_db` - Transaction service
- `fraud_db` - Fraud service (when implemented)

## Redis Access

Connect to Redis:
```bash
docker exec -it bankflow-redis redis-cli -a redis123
```

## Kafka Topics

Kafka automatically creates topics when first used. Common topics:
- `identity-events`
- `customer-events`
- `account-events`
- `transaction-events`

List topics:
```bash
docker exec -it bankflow-kafka kafka-topics --bootstrap-server localhost:9092 --list
```

## Environment Variables

All environment variables are set in `docker-compose.yml`. To override for local development:

1. Copy `docker-compose.override.yml.example` to `docker-compose.override.yml`
2. Modify as needed
3. Run `docker-compose up`

## Development Workflow

### Rebuild a Single Service

```bash
docker compose build identity-service
docker compose up -d identity-service
```

### View Service Logs

```bash
docker compose logs -f identity-service
```

### Execute Commands in Container

```bash
# Shell into a service container
docker exec -it bankflow-identity-service sh

# Run a command
docker exec -it bankflow-customer-service java -version
```

### Hot Reload (Development)

For hot-reload during development, mount your source code:

1. Create `docker-compose.override.yml`:
```yaml
version: '3.8'
services:
  identity-service:
    volumes:
      - ./services/identity-service:/app
```

2. Use development tools inside containers or run services locally and connect to Docker infrastructure.

## Troubleshooting

### Services Won't Start

1. **Check logs**: `docker-compose logs <service-name>`
2. **Check health**: `docker-compose ps`
3. **Verify dependencies**: Ensure PostgreSQL, Redis, and Kafka are healthy
4. **Rebuild**: `docker-compose build --no-cache <service-name>`

### Database Connection Issues

1. Ensure PostgreSQL is healthy: `docker-compose ps postgres`
2. Check if databases were created: `docker exec -it bankflow-postgres psql -U bankflow -l`
3. Re-run init script if needed: `docker exec -i bankflow-postgres psql -U bankflow < scripts/init-databases.sql`

### Kafka Connection Issues

1. Check Kafka health: `docker-compose ps kafka`
2. Verify Kafka is accessible: `docker exec -it bankflow-kafka kafka-broker-api-versions --bootstrap-server localhost:9092`
3. Check Zookeeper: `docker-compose ps zookeeper`

### Port Conflicts

If ports are already in use, modify ports in `docker-compose.yml`:
```yaml
ports:
  - "8001:8001"  # Change first number to available port
```

### Clean Start

To completely reset everything:
```bash
docker compose down -v
docker system prune -f
docker compose up --build
```

## Testing

### Health Checks

```bash
# Identity Service
curl http://localhost:8001/health

# Customer Service
curl http://localhost:8002/health

# Account Service
curl http://localhost:8004/health

# Transaction Service
curl http://localhost:8003/health
```

### API Testing

```bash
# Register a user
curl -X POST http://localhost:8001/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "phone": "+1234567890"
  }'
```

## Production Considerations

⚠️ **This setup is for development only!** For production, use infrastructure

## Next Steps

- See `proto/README.md` for gRPC setup
- See `proto/GENERATION_GUIDE.md` for proto code generation
- See individual service READMEs for service-specific documentation

