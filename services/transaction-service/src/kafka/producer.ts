import { Kafka, Producer, ProducerRecord } from 'kafkajs';

export class KafkaProducer {
  private readonly kafka: Kafka;
  private readonly producer: Producer;
  private isConnected: boolean = false;

  constructor(brokers: string[]) {
    this.kafka = new Kafka({
      clientId: 'transaction-service',
      brokers,
    });
    this.producer = this.kafka.producer();
  }

  async connect(): Promise<void> {
    if (!this.isConnected) {
      await this.producer.connect();
      this.isConnected = true;
      console.log('Kafka producer connected');
    }
  }

  async disconnect(): Promise<void> {
    if (this.isConnected) {
      await this.producer.disconnect();
      this.isConnected = false;
      console.log('Kafka producer disconnected');
    }
  }

  async publish(topic: string, message: any): Promise<void> {
    try {
      if (!this.isConnected) {
        await this.connect();
      }

      const record: ProducerRecord = {
        topic,
        messages: [
          {
            value: JSON.stringify(message),
            timestamp: Date.now().toString(),
          },
        ],
      };

      await this.producer.send(record);
      console.log(`Message published to topic ${topic}:`, message.event_type);
    } catch (error) {
      console.error('Error publishing message:', error);
      throw error;
    }
  }
}
