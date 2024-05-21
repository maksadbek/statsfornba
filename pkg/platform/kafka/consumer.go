package kafka

import (
	"context"

	"github.com/IBM/sarama"
)

type MessageHandler func(key, value []byte) error

type Consumer struct {
	handler MessageHandler
	client  sarama.ConsumerGroup

	topics []string
	done   chan struct{}
}

func NewConsumer(brokers, topics []string, group, clientID string, handler MessageHandler) (*Consumer, error) {
	conf := sarama.NewConfig()

	conf.Metadata.Full = true
	conf.Version = sarama.V3_3_1_0
	conf.ClientID = clientID
	conf.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}

	client, err := sarama.NewConsumerGroup(brokers, group, conf)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		client:  client,
		handler: handler,
		done:    make(chan struct{}),
		topics:  topics,
	}, nil
}

func (c *Consumer) Consume(ctx context.Context) error {
	groupHandler := ConsumerGroupHandler{
		handler: c.handler,
	}

	for {
		select {
		case <-ctx.Done():
			c.done <- struct{}{}
			return context.Canceled
		default:
		}

		err := c.client.Consume(ctx, c.topics, &groupHandler)
		if err != nil {
			return err
		}

		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
	}
}

func (c *Consumer) WaitUntilShutdown(ctx context.Context) {
	// close the reader when there are consumers left.
	defer c.client.Close() //nolint:errcheck

	select {
	case <-ctx.Done():
	case <-c.done:
	}
}

// ConsumerGroupHandler represents a Sarama consumer group consumer
type ConsumerGroupHandler struct {
	handler MessageHandler
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (*ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			err := c.handler(message.Key, message.Value)
			if err != nil {
				continue
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (*ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}
