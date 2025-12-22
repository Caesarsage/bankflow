export enum TransactionType {
  TRANSFER = 'TRANSFER',
  DEPOSIT = 'DEPOSIT',
  WITHDRAWAL = 'WITHDRAWAL',
  PAYMENT = 'PAYMENT',
  REFUND = 'REFUND',
  FEE = 'FEE',
}

export enum TransactionStatus {
  PENDING = 'PENDING',
  PROCESSING = 'PROCESSING',
  COMPLETED = 'COMPLETED',
  FAILED = 'FAILED',
  REVERSED = 'REVERSED',
  CANCELLED = 'CANCELLED',
}

export interface Transaction {
  id: string;
  transactionRef: string;
  fromAccountId?: string;
  toAccountId?: string;
  amount: number;
  currency: string;
  transactionType: TransactionType;
  status: TransactionStatus;
  description?: string;
  metadata?: Record<string, any>;
  processedAt?: Date;
  reversedAt?: Date;
  createdAt: Date;
  updatedAt: Date;
}

export interface CreateTransferRequest {
  fromAccountId: string;
  toAccountId: string;
  amount: number;
  currency: string;
  description?: string;
  metadata?: Record<string, any>;
}

export interface CreatePaymentRequest {
  fromAccountId: string;
  toAccountId: string;
  amount: number;
  currency: string;
  paymentMethod: string;
  description?: string;
  metadata?: Record<string, any>;
}

export interface TransactionSearchParams {
  accountId?: string;
  fromDate?: Date;
  toDate?: Date;
  status?: TransactionStatus;
  type?: TransactionType;
  minAmount?: number;
  maxAmount?: number;
  page?: number;
  limit?: number;
}

export interface TransactionSummary {
  totalTransactions: number;
  totalAmount: number;
  successfulTransactions: number;
  failedTransactions: number;
  pendingTransactions: number;
}

export interface ScheduledTransfer {
  id: string;
  fromAccountId: string;
  toAccountId: string;
  amount: number;
  currency: string;
  frequency: 'ONCE' | 'DAILY' | 'WEEKLY' | 'MONTHLY';
  startDate: Date;
  endDate?: Date;
  nextExecutionDate: Date;
  description?: string;
  isActive: boolean;
  createdAt: Date;
  updatedAt: Date;
}
