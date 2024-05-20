package controller

import (
	"encoding/json"

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

	return err
}
