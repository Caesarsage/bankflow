import { Request, Response } from 'express';
import { TransactionService } from '../services/transaction.service';
import { CreateTransferRequest, CreatePaymentRequest, TransactionSearchParams, TransactionStatus, TransactionType } from '../models/transaction.model';

export class TransactionController {

  constructor(private readonly transactionService: TransactionService) { }

  createTransfer = async (req: Request, res: Response): Promise<void> => {
    try {
      const transferReq: CreateTransferRequest = req.body;

      // Validate request
      if (!transferReq.fromAccountId || !transferReq.toAccountId || !transferReq.amount) {
        res.status(400).json({ error: 'Missing required fields' });
        return;
      }

      if (transferReq.amount <= 0) {
        res.status(400).json({ error: 'Amount must be greater than zero' });
        return;
      }

      const transaction = await this.transactionService.createTransfer(transferReq);
      res.status(201).json(transaction);
    } catch (error) {
      console.error('Error creating transfer:', error);
      res.status(500).json({ error: 'Failed to create transfer' });
    }
  };

  createPayment = async (req: Request, res: Response): Promise<void> => {
    try {
      const paymentReq: CreatePaymentRequest = req.body;

      // Validate request
      if (!paymentReq.fromAccountId || !paymentReq.toAccountId || !paymentReq.amount) {
        res.status(400).json({ error: 'Missing required fields' });
        return;
      }

      // For now, treat payment same as transfer
      const transferReq: CreateTransferRequest = {
        fromAccountId: paymentReq.fromAccountId,
        toAccountId: paymentReq.toAccountId,
        amount: paymentReq.amount,
        currency: paymentReq.currency,
        description: paymentReq.description,
        metadata: {
          ...paymentReq.metadata,
          paymentMethod: paymentReq.paymentMethod,
        },
      };

      const transaction = await this.transactionService.createTransfer(transferReq);
      res.status(201).json(transaction);
    } catch (error) {
      console.error('Error creating payment:', error);
      res.status(500).json({ error: 'Failed to create payment' });
    }
  };

  getTransaction = async (req: Request, res: Response): Promise<void> => {
    try {
      const { id } = req.params;
      const transaction = await this.transactionService.getTransactionById(id);

      if (!transaction) {
        res.status(404).json({ error: 'Transaction not found' });
        return;
      }

      res.json(transaction);
    } catch (error) {
      console.error('Error getting transaction:', error);
      res.status(500).json({ error: 'Failed to get transaction' });
    }
  };

  getAccountTransactions = async (req: Request, res: Response): Promise<void> => {
    try {
      const { accountId } = req.params;
      const { status, type, fromDate, toDate, page, limit } = req.query;

      const params: TransactionSearchParams = {
        status: status as TransactionStatus | undefined,
        type: type as TransactionType | undefined,
        fromDate: fromDate ? new Date(fromDate as string) : undefined,
        toDate: toDate ? new Date(toDate as string) : undefined,
        page: page ? Number.parseInt(page as string) : 1,
        limit: limit ? Number.parseInt(limit as string) : 50,
      };

      const transactions = await this.transactionService.getTransactionsByAccount(
        accountId,
        params
      );

      res.json({
        transactions,
        page: params.page,
        limit: params.limit,
        total: transactions.length,
      });
    } catch (error) {
      console.error('Error getting account transactions:', error);
      res.status(500).json({ error: 'Failed to get transactions' });
    }
  };

  searchTransactions = async (req: Request, res: Response): Promise<void> => {
    try {
      const { accountId, status, type, fromDate, toDate, minAmount, maxAmount, page, limit } = req.query;

      if (!accountId) {
        res.status(400).json({ error: 'Account ID is required' });
        return;
      }

      const params: TransactionSearchParams = {
        status: status as TransactionStatus | undefined,
        type: type as TransactionType | undefined,
        fromDate: fromDate ? new Date(fromDate as string) : undefined,
        toDate: toDate ? new Date(toDate as string) : undefined,
        minAmount: minAmount ? Number.parseFloat(minAmount as string) : undefined,
        maxAmount: maxAmount ? Number.parseFloat(maxAmount as string) : undefined,
        page: page ? Number.parseInt(page as string) : 1,
        limit: limit ? Number.parseInt(limit as string) : 50,
      };

      const transactions = await this.transactionService.getTransactionsByAccount(
        accountId as string,
        params
      );

      res.json({
        transactions,
        page: params.page,
        limit: params.limit,
        total: transactions.length,
      });
    } catch (error) {
      console.error('Error searching transactions:', error);
      res.status(500).json({ error: 'Failed to search transactions' });
    }
  };

  reverseTransaction = async (req: Request, res: Response): Promise<void> => {
    try {
      const { id } = req.params;

      const reversalTransaction = await this.transactionService.reverseTransaction(id);
      res.json({
        message: 'Transaction reversed successfully',
        reversal: reversalTransaction,
      });
    } catch (error: any) {
      console.error('Error reversing transaction:', error);
      res.status(400).json({ error: error.message || 'Failed to reverse transaction' });
    }
  };
}
