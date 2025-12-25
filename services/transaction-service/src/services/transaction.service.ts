import { Pool } from 'pg';
import { v4 as uuidv4 } from 'uuid';
import { KafkaProducer } from '../kafka/producer';
import { getAccountServiceClient } from '../clients/account.client';
import {
  Transaction,
  CreateTransferRequest,
  TransactionType,
  TransactionStatus,
  TransactionSearchParams,
} from '../models/transaction.model';

export class TransactionService {
  constructor(
    private readonly db: Pool,
    private readonly kafkaProducer: KafkaProducer
  ) { }

  async createTransfer(req: CreateTransferRequest): Promise<Transaction> {
    const client = await this.db.connect();

    try {
      await client.query('BEGIN');

      // Generate transaction reference
      const transactionRef = this.generateTransactionRef();

      // Create transaction record
      const transaction: Transaction = {
        id: uuidv4(),
        transactionRef,
        fromAccountId: req.fromAccountId,
        toAccountId: req.toAccountId,
        amount: req.amount,
        currency: req.currency,
        transactionType: TransactionType.TRANSFER,
        status: TransactionStatus.PENDING,
        description: req.description,
        metadata: req.metadata,
        createdAt: new Date(),
        updatedAt: new Date(),
      };

      const insertQuery = `
        INSERT INTO transactions (
          id, transaction_ref, from_account_id, to_account_id,
          amount, currency, transaction_type, status,
          description, metadata, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        RETURNING *
      `;

      const result = await client.query(insertQuery, [
        transaction.id,
        transaction.transactionRef,
        transaction.fromAccountId,
        transaction.toAccountId,
        transaction.amount,
        transaction.currency,
        transaction.transactionType,
        transaction.status,
        transaction.description,
        JSON.stringify(transaction.metadata || {}),
        transaction.createdAt,
        transaction.updatedAt,
      ]);

      await client.query('COMMIT');

      const createdTransaction = this.mapRowToTransaction(result.rows[0]);

      // Publish event
      await this.publishTransactionInitiated(createdTransaction);

      // Process transaction asynchronously
      this.processTransfer(createdTransaction).catch((err) => {
        console.error('Error processing transfer:', err);
      });

      return createdTransaction;
    } catch (error) {
      await client.query('ROLLBACK');
      throw error;
    } finally {
      client.release();
    }
  }

  async processTransfer(transaction: Transaction): Promise<void> {
    try {
      // Update status to processing
      await this.updateTransactionStatus(transaction.id, TransactionStatus.PROCESSING);
      await this.publishTransactionProcessing(transaction);

      // Validate balance before processing
      await this.validateAccountBalance(transaction.fromAccountId!, transaction.amount);

      // Call account service to debit from account
      await this.debitAccount(transaction.fromAccountId!, transaction.amount, transaction.transactionRef);

      // Credit to account
      await this.creditAccount(transaction.toAccountId!, transaction.amount, transaction.transactionRef);

      // Update status to completed
      await this.updateTransactionStatus(transaction.id, TransactionStatus.COMPLETED);
      await this.markTransactionProcessed(transaction.id);
      await this.publishTransactionCompleted(transaction);

    } catch (error) {
      console.error('Transfer processing failed:', error);
      await this.updateTransactionStatus(transaction.id, TransactionStatus.FAILED);
      await this.publishTransactionFailed(transaction);
      throw error;
    }
  }

  async getTransactionById(id: string): Promise<Transaction | null> {
    const query = 'SELECT * FROM transactions WHERE id = $1';
    const result = await this.db.query(query, [id]);

    if (result.rows.length === 0) {
      return null;
    }

    return this.mapRowToTransaction(result.rows[0]);
  }

  async getTransactionsByAccount(
    accountId: string,
    params: TransactionSearchParams = {}
  ): Promise<Transaction[]> {
    let query = `
      SELECT * FROM transactions
      WHERE (from_account_id = $1 OR to_account_id = $1)
    `;
    const queryParams: any[] = [accountId];
    let paramIndex = 2;

    if (params.status) {
      query += ` AND status = $${paramIndex}`;
      queryParams.push(params.status);
      paramIndex++;
    }

    if (params.type) {
      query += ` AND transaction_type = $${paramIndex}`;
      queryParams.push(params.type);
      paramIndex++;
    }

    if (params.fromDate) {
      query += ` AND created_at >= $${paramIndex}`;
      queryParams.push(params.fromDate);
      paramIndex++;
    }

    if (params.toDate) {
      query += ` AND created_at <= $${paramIndex}`;
      queryParams.push(params.toDate);
      paramIndex++;
    }

    query += ' ORDER BY created_at DESC';

    const limit = params.limit || 50;
    const page = params.page || 1;
    const offset = (page - 1) * limit;

    query += ` LIMIT $${paramIndex} OFFSET $${paramIndex + 1}`;
    queryParams.push(limit, offset);

    const result = await this.db.query(query, queryParams);
    return result.rows.map(this.mapRowToTransaction);
  }

  async reverseTransaction(transactionId: string): Promise<Transaction> {
    const client = await this.db.connect();

    try {
      await client.query('BEGIN');

      // Get original transaction
      const original = await this.getTransactionById(transactionId);
      if (!original) {
        throw new Error('Transaction not found');
      }

      if (original.status !== TransactionStatus.COMPLETED) {
        throw new Error('Can only reverse completed transactions');
      }

      // Create reversal transaction
      const reversalTransaction: Transaction = {
        id: uuidv4(),
        transactionRef: this.generateTransactionRef(),
        fromAccountId: original.toAccountId,
        toAccountId: original.fromAccountId,
        amount: original.amount,
        currency: original.currency,
        transactionType: TransactionType.REFUND,
        status: TransactionStatus.COMPLETED,
        description: `Reversal of ${original.transactionRef}`,
        metadata: { originalTransactionId: original.id },
        processedAt: new Date(),
        createdAt: new Date(),
        updatedAt: new Date(),
      };

      const insertQuery = `
        INSERT INTO transactions (
          id, transaction_ref, from_account_id, to_account_id,
          amount, currency, transaction_type, status,
          description, metadata, processed_at, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
        RETURNING *
      `;

      await client.query(insertQuery, [
        reversalTransaction.id,
        reversalTransaction.transactionRef,
        reversalTransaction.fromAccountId,
        reversalTransaction.toAccountId,
        reversalTransaction.amount,
        reversalTransaction.currency,
        reversalTransaction.transactionType,
        reversalTransaction.status,
        reversalTransaction.description,
        JSON.stringify(reversalTransaction.metadata),
        reversalTransaction.processedAt,
        reversalTransaction.createdAt,
        reversalTransaction.updatedAt,
      ]);

      // Mark original as reversed
      await client.query(
        'UPDATE transactions SET status = $1, reversed_at = $2 WHERE id = $3',
        [TransactionStatus.REVERSED, new Date(), transactionId]
      );

      await client.query('COMMIT');

      // Update account balances
      await this.debitAccount(original.toAccountId!, original.amount, original.transactionRef);
      await this.creditAccount(original.fromAccountId!, original.amount, original.transactionRef);

      // Publish event
      await this.publishTransactionReversed(original);

      return reversalTransaction;
    } catch (error) {
      await client.query('ROLLBACK');
      throw error;
    } finally {
      client.release();
    }
  }

  private async updateTransactionStatus(id: string, status: TransactionStatus): Promise<void> {
    const query = 'UPDATE transactions SET status = $1, updated_at = $2 WHERE id = $3';
    await this.db.query(query, [status, new Date(), id]);
  }

  private async markTransactionProcessed(id: string): Promise<void> {
    const query = 'UPDATE transactions SET processed_at = $1, updated_at = $2 WHERE id = $3';
    await this.db.query(query, [new Date(), new Date(), id]);
  }

  private generateTransactionRef(): string {
    const timestamp = Date.now().toString(36).toUpperCase();
    const random = Math.random().toString(36).substring(2, 8).toUpperCase();
    return `TXN-${timestamp}-${random}`;
  }

  private mapRowToTransaction(row: any): Transaction {
    return {
      id: row.id,
      transactionRef: row.transaction_ref,
      fromAccountId: row.from_account_id,
      toAccountId: row.to_account_id,
      amount: parseFloat(row.amount),
      currency: row.currency,
      transactionType: row.transaction_type,
      status: row.status,
      description: row.description,
      metadata: row.metadata,
      processedAt: row.processed_at,
      reversedAt: row.reversed_at,
      createdAt: row.created_at,
      updatedAt: row.updated_at,
    };
  }

  // Account service interactions via HTTP
  private async debitAccount(accountId: string, amount: number, transactionRef: string): Promise<void> {
    const accountClient = getAccountServiceClient();
    await accountClient.debitAccount(accountId, amount, transactionRef);
  }

  private async creditAccount(accountId: string, amount: number, transactionRef: string): Promise<void> {
    const accountClient = getAccountServiceClient();
    await accountClient.creditAccount(accountId, amount, transactionRef);
  }

  private async validateAccountBalance(accountId: string, amount: number): Promise<void> {
    const accountClient = getAccountServiceClient();
    const hasFunds = await accountClient.hasSufficientBalance(accountId, amount);
    if (!hasFunds) {
      throw new Error('Insufficient funds');
    }
  }

  // Kafka event publishing
  private async publishTransactionInitiated(transaction: Transaction): Promise<void> {
    await this.kafkaProducer.publish('transaction-events', {
      event_type: 'transaction.initiated',
      transaction_id: transaction.id,
      transaction_ref: transaction.transactionRef,
      from_account_id: transaction.fromAccountId,
      to_account_id: transaction.toAccountId,
      amount: transaction.amount,
      currency: transaction.currency,
      timestamp: Date.now(),
    });
  }

  private async publishTransactionProcessing(transaction: Transaction): Promise<void> {
    await this.kafkaProducer.publish('transaction-events', {
      event_type: 'transaction.processing',
      transaction_id: transaction.id,
      transaction_ref: transaction.transactionRef,
      timestamp: Date.now(),
    });
  }

  private async publishTransactionCompleted(transaction: Transaction): Promise<void> {
    await this.kafkaProducer.publish('transaction-events', {
      event_type: 'transaction.completed',
      transaction_id: transaction.id,
      transaction_ref: transaction.transactionRef,
      from_account_id: transaction.fromAccountId,
      to_account_id: transaction.toAccountId,
      amount: transaction.amount,
      currency: transaction.currency,
      timestamp: Date.now(),
    });
  }

  private async publishTransactionFailed(transaction: Transaction): Promise<void> {
    await this.kafkaProducer.publish('transaction-events', {
      event_type: 'transaction.failed',
      transaction_id: transaction.id,
      transaction_ref: transaction.transactionRef,
      timestamp: Date.now(),
    });
  }

  private async publishTransactionReversed(transaction: Transaction): Promise<void> {
    await this.kafkaProducer.publish('transaction-events', {
      event_type: 'transaction.reversed',
      transaction_id: transaction.id,
      transaction_ref: transaction.transactionRef,
      timestamp: Date.now(),
    });
  }
}
