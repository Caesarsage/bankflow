import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';
import path from 'path';

const PROTO_PATH = path.join(__dirname, '../../proto/account/account.proto');

const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
});

const accountProto = grpc.loadPackageDefinition(packageDefinition).account as any;

export class AccountGRPCClient {
  private client: any;

  constructor(serverAddress: string = 'localhost:50051') {
    this.client = new accountProto.AccountService(
      serverAddress,
      grpc.credentials.createInsecure()
    );
  }

  async getAccount(accountId: string): Promise<any> {
    return new Promise((resolve, reject) => {
      this.client.GetAccount({ account_id: accountId }, (error: any, response: any) => {
        if (error) {
          reject(error);
        } else if (response.error) {
          reject(new Error(response.error.message));
        } else {
          resolve(response.account);
        }
      });
    });
  }

  async getBalance(accountId: string): Promise<{ balance: number; available_balance: number }> {
    return new Promise((resolve, reject) => {
      this.client.GetBalance({ account_id: accountId }, (error: any, response: any) => {
        if (error) {
          reject(error);
        } else if (response.error) {
          reject(new Error(response.error.message));
        } else {
          resolve({
            balance: response.balance,
            available_balance: response.available_balance,
          });
        }
      });
    });
  }

  async updateBalance(accountId: string, amount: number, transactionRef: string): Promise<number> {
    return new Promise((resolve, reject) => {
      this.client.UpdateBalance(
        {
          account_id: accountId,
          amount,
          transaction_ref: transactionRef,
        },
        (error: any, response: any) => {
          if (error) {
            reject(error);
          } else if (!response.success) {
            reject(new Error(response.error?.message || 'Update failed'));
          } else {
            resolve(response.new_balance);
          }
        }
      );
    });
  }

  async createHold(
    accountId: string,
    amount: number,
    reason: string,
    transactionRef?: string
  ): Promise<string> {
    return new Promise((resolve, reject) => {
      this.client.CreateHold(
        {
          account_id: accountId,
          amount,
          reason,
          transaction_ref: transactionRef || '',
        },
        (error: any, response: any) => {
          if (error) {
            reject(error);
          } else if (!response.success) {
            reject(new Error(response.error?.message || 'Create hold failed'));
          } else {
            resolve(response.hold_id);
          }
        }
      );
    });
  }

  async releaseHold(holdId: string): Promise<void> {
    return new Promise((resolve, reject) => {
      this.client.ReleaseHold({ hold_id: holdId }, (error: any, response: any) => {
        if (error) {
          reject(error);
        } else if (!response.success) {
          reject(new Error(response.error?.message || 'Release hold failed'));
        } else {
          resolve();
        }
      });
    });
  }

  // Server streaming example
  streamBalanceUpdates(accountId: string, callback: (update: any) => void): grpc.ClientReadableStream<any> {
    const stream = this.client.StreamBalanceUpdates({ account_id: accountId });

    stream.on('data', (update: any) => {
      callback(update);
    });

    stream.on('error', (error: any) => {
      console.error('Stream error:', error);
    });

    stream.on('end', () => {
      console.log('Stream ended');
    });

    return stream;
  }

  close(): void {
    grpc.closeClient(this.client);
  }
}
