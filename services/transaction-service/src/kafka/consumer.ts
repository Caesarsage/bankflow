import { Kafka, Consumer, EachMessagePayload } from 'kafkajs';

export class KafkaConsumer {
  private readonly kafka: Kafka;
  private readonly consumer: Consumer;
  private isConnected: boolean = false;

  constructor(brokers: string[], groupId: string) {
    this.kafka = new Kafka({
      clientId: 'transaction-service',
      brokers,
    });
    this.consumer = this.kafka.consumer({ groupId });
  }

  async connect(): Promise<void> {
    if (!this.isConnected) {
      await this.consumer.connect();
      this.isConnected = true;
      console.log('Kafka consumer connected');
    }
  }

  async disconnect(): Promise<void> {
    if (this.isConnected) {
      await this.consumer.disconnect();
      this.isConnected = false;
      console.log('Kafka consumer disconnected');
    }
  }

  async subscribe(topics: string[]): Promise<void> {
    if (!this.isConnected) {
      await this.connect();
    }

    for (const topic of topics) {
      await this.consumer.subscribe({ topic, fromBeginning: false });
      console.log(`Subscribed to topic: ${topic}`);
    }
  }

  async consume(handler: (message: any) => Promise<void>): Promise<void> {
    await this.consumer.run({
      eachMessage: async ({ topic, message }: EachMessagePayload) => {
        try {
          const value = message.value?.toString();
          if (value) {
            const parsedMessage = JSON.parse(value);
            console.log(`Received message from ${topic}:`, parsedMessage.event_type);
            await handler(parsedMessage);
          }
        } catch (error) {
          console.error('Error processing message:', error);
        }
      },
    });
  }
}
