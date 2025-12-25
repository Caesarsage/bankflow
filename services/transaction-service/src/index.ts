import express, { Express, Request, Response } from 'express';
import cors from 'cors';
import helmet from 'helmet';
import { Pool } from 'pg';
import { KafkaProducer } from './kafka/producer';
import { createAccountServiceClient } from './clients/account.client';
import { TransactionService } from './services/transaction.service';
import { TransactionController } from './controllers/transaction.controller';
import { createTransactionRoutes } from './routes/transaction.routes';

const app: Express = express();
const PORT = process.env.PORT || 8003;

// Middleware
app.use(helmet());
app.use(cors());
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// Database connection
const pool = new Pool({
  host: process.env.DB_HOST || 'localhost',
  port: Number.parseInt(process.env.DB_PORT || '5432'),
  database: process.env.DB_NAME || 'transactions',
  user: process.env.DB_USER || 'postgres',
  password: process.env.DB_PASSWORD || 'postgres',
  max: 20,
  idleTimeoutMillis: 30000,
  connectionTimeoutMillis: 2000,
});

// Kafka configuration
const kafkaBrokers = process.env.KAFKA_BROKERS?.split(',') || ['localhost:9092'];
const kafkaProducer = new KafkaProducer(kafkaBrokers);

// Initialize services
let transactionService: TransactionService;
let transactionController: any;

async function initializeServices() {
  try {
    // Test database connection
    await pool.query('SELECT NOW()');
    console.log('Connected to database successfully');

    // Connect Kafka producer
    await kafkaProducer.connect();

    // Initialize Account Service client (gRPC)
    const accountServiceURL = process.env.ACCOUNT_SERVICE_GRPC_URL || 'account-service:50051';
    createAccountServiceClient(accountServiceURL);
    console.log(`Account Service gRPC client initialized: ${accountServiceURL}`);

    // Initialize service and controller
    transactionService = new TransactionService(pool, kafkaProducer);
    transactionController = new TransactionController(transactionService);

    console.log('Services initialized successfully');
  } catch (error) {
    console.error('Failed to initialize services:', error);
    process.exit(1);
  }
}

// Health check
app.get('/health', (_req: Request, res: Response) => {
  res.json({
    status: 'healthy',
    service: 'transaction-service',
    timestamp: new Date().toISOString(),
  });
});

// API routes
app.use('/api/v1/transactions', (_req, res, next) => {
  if (!transactionController) {
    res.status(503).json({ error: 'Service not ready' });
    return;
  }
  next();
}, createTransactionRoutes(transactionController));

// Error handling middleware
app.use((err: Error, _req: Request, res: Response, _next: any) => {
  console.error('Error:', err);
  res.status(500).json({
    error: 'Internal server error',
    message: err.message,
  });
});

// 404 handler
app.use((_req: Request, res: Response) => {
  res.status(404).json({ error: 'Not found' });
});

// Start server
async function startServer() {
  await initializeServices();

  app.listen(PORT, () => {
    console.log(`Transaction service listening on port ${PORT}`);
  });
}

// Graceful shutdown
process.on('SIGTERM', async () => {
  console.log('SIGTERM signal received: closing HTTP server');
  await kafkaProducer.disconnect();
  await pool.end();
  process.exit(0);
});

process.on('SIGINT', async () => {
  console.log('SIGINT signal received: closing HTTP server');
  await kafkaProducer.disconnect();
  await pool.end();
  process.exit(0);
});

// Start the server
startServer().catch((error) => {
  console.error('Failed to start server:', error);
  process.exit(1);
});

export default app;
