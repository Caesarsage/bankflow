import { Router } from 'express';
import { TransactionController } from '../controllers/transaction.controller';

export function createTransactionRoutes(controller: TransactionController): Router {
  const router = Router();

  // Create transfer
  router.post('/transfer', controller.createTransfer);

  // Create payment
  router.post('/payment', controller.createPayment);

  // Get transaction by ID
  router.get('/:id', controller.getTransaction);

  // Get account transactions
  router.get('/account/:accountId', controller.getAccountTransactions);

  // Search transactions
  router.get('/search', controller.searchTransactions);

  // Reverse transaction
  router.post('/:id/reverse', controller.reverseTransaction);

  return router;
}
