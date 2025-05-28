package messaging

import (
	"context"

	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/logger"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaProducer struct {
	producer *kafka.Producer
}

type KafkaConsumer struct {
	consumer *kafka.Consumer
}

func NewKafkaProducer(brokers string) (*KafkaProducer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
	})
	if err != nil {
		return nil, err
	}
	return &KafkaProducer{producer: p}, nil
}

func (p *KafkaProducer) Produce(ctx context.Context, topic, key string, value []byte) error {
	deliveryChan := make(chan kafka.Event)
	defer close(deliveryChan)

	err := p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          value,
	}, deliveryChan)
	if err != nil {
		return err
	}

	e := <-deliveryChan
	m, ok := e.(*kafka.Message)
	if !ok || m.TopicPartition.Error != nil {
		return m.TopicPartition.Error
	}
	logger.Info("Message delivered", "topic", topic, "key", key)
	return nil
}

func NewKafkaConsumer(brokers, groupID string, topics []string) (*KafkaConsumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}
	err = c.SubscribeTopics(topics, nil)
	if err != nil {
		return nil, err
	}
	return &KafkaConsumer{consumer: c}, nil
}

func (c *KafkaConsumer) Consume(ctx context.Context, handler func(message *kafka.Message)) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := c.consumer.ReadMessage(-1)
			if err != nil {
				logger.Error("Consumer error", "error", err)
				continue
			}
			handler(msg)
		}
	}
}
