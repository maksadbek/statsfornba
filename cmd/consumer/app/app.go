package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	"go.uber.org/ratelimit"

	"github.com/maksadbek/statsfornba/internal/model"
	repository "github.com/maksadbek/statsfornba/internal/repository/postgres"
	"github.com/maksadbek/statsfornba/pkg/platform/kafka"
)

type Config struct {
	Kafka struct {
		Addr              []string `required:"true"`
		Topic             []string `required:"true"`
		ConsumerRateLimit int      `default:"100"`
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

	rl := ratelimit.New(c.Kafka.ConsumerRateLimit)

	consumer, err := kafka.NewConsumer(c.Kafka.Addr, c.Kafka.Topic, "stats-consumer-group", "consumer-app", kafka.MessageHandler(func(key, value []byte) error {
		log.Printf("recevied message, key = %v, value = %v", string(key), string(value))

		var stat model.Stat

		err := json.Unmarshal(value, &stat)
		if err != nil {
			log.Println("invalid payload, unable to unmarshal", err)
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		rl.Take() // wait until ratelimiter allows to proceed

		err = statsRepo.Add(ctx, &stat)
		if err != nil {
			log.Println("failed to create stat", err)
			return err
		}

		return nil
	}))

	if err != nil {
		return err
	}

	cooldownPeriod := time.Minute

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			err = consumer.Consume(ctx)
			if err != nil {
				if err == context.Canceled {
					log.Println("context cancelled, stopping!")
					return
				}

				log.Println("failed to start Kafka consumer", err)
				time.Sleep(cooldownPeriod)
			} else {
				return
			}
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)

	<-shutdown

	log.Println("shutting down application")

	cancel()

	ctx, cancel = context.WithTimeout(ctx, time.Minute)
	defer cancel()

	log.Println("wait until consumers shutdown")
	consumer.WaitUntilShutdown(ctx)

	return nil
}
