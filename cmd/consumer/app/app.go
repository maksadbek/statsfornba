package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"

	"github.com/maksadbek/statsfornba/internal/model"
	repository "github.com/maksadbek/statsfornba/internal/repository/postgres"
	"github.com/maksadbek/statsfornba/pkg/platform/kafka"
)

type Config struct {
	Kafka struct {
		Addr  []string `required:"true"`
		Topic []string `required:"true"`
	}

	Postgres struct {
		DSN string `required:"true"`
	}
}

func Run() error {
	var c Config
	err := envconfig.Process("consumer", &c)
	if err != nil {
		return err
	}

	conn, err := sql.Open("postgres", c.Postgres.DSN)
	if err != nil {
		return err
	}

	statsRepo, err := repository.NewStats(conn)
	if err != nil {
		return err
	}

	consumer, err := kafka.NewConsumer(c.Kafka.Addr, c.Kafka.Topic, "stats-consumer-group", "consumer-app", kafka.MessageHandler(func(key, value []byte) error {
		fmt.Println(string(key), string(value))

		var stat model.Stat

		err := json.Unmarshal(value, &stat)
		if err != nil {
			fmt.Println("error", err)
		}

		err = statsRepo.Add(context.Background(), &stat)
		if err != nil {
			fmt.Println(err)
		}

		return nil
	}))
	if err != nil {
		return err
	}

	err = consumer.Consume(context.Background())

	return err
}
