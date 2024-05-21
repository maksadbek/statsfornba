package controller

import (
	"encoding/json"
	"fmt"

	"github.com/maksadbek/statsfornba/internal/model"
	"github.com/maksadbek/statsfornba/pkg/platform/kafka"
)

type Stats struct {
	publisher *kafka.Publisher
}

func NewStats(p *kafka.Publisher) *Stats {
	return &Stats{
		publisher: p,
	}
}

func (c *Stats) Add(stat *model.Stat) error {
	js, err := json.Marshal(stat)
	if err != nil {
		return err
	}

	err = c.publisher.Publish(string(js))
	if err != nil {
		return fmt.Errorf("failed to publish the event to Kafka: %w", err)
	}

	return nil
}
