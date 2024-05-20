package kafka

import (
	"log"

	"github.com/IBM/sarama"
)

type Publisher struct {
	topic string

	publisher sarama.SyncProducer
}

func NewPublisher(brokers []string, topic, clientID string) (*Publisher, error) {
	conf := sarama.NewConfig()

	conf.Metadata.Full = true
	conf.Version = sarama.V3_3_1_0
	conf.ClientID = clientID
	conf.Producer.RequiredAcks = sarama.WaitForAll
	conf.Producer.Retry.Max = 10
	conf.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, conf)
	if err != nil {
		log.Fatalln("Failed to start Sarama producer:", err)
	}

	return &Publisher{
		publisher: producer,
		topic:     topic,
	}, nil
}

func (p *Publisher) Publish(message string) error {
	_, _, err := p.publisher.SendMessage(&sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.StringEncoder(message),
	})

	return err
}
