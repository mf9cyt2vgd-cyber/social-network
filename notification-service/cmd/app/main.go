package main

import (
	"context"
	"fmt"
	"notifications-service/internal/config"
	"notifications-service/internal/logger"
	"notifications-service/internal/mapper"
	"notifications-service/internal/transport/kafka"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	cfg := config.MustLoad()
	log := logger.SetupLogger(cfg.Env)
	ctxC, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := kafka.NewKafkaConsumer(cfg.Brokers, cfg.GroupID, cfg.Topic, log)
	if err != nil {
		log.Error("failed to create kafka", "error", err)
	}

	wg := new(sync.WaitGroup)
	errChan := make(chan error)
	msgschan := make(chan []byte)
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	wg.Go(func() {
		c.Consume(ctxC, msgschan, errChan)
	})
	for {
		select {
		case msg := <-msgschan:
			post, err := mapper.ConvertKafkaMessageIntoPost(msg)
			if err != nil {
				log.Error("failed to convert message", "error", err)
				continue
			}
			log.Info("successfully read post from Kafka", "post.id", post.ID)
			fmt.Println(post)
		case e := <-errChan:
			log.Error("got error from kafka", "error", e)
			continue
		case <-done:
			log.Info("got interrupt signal")
			cancel()
			wg.Wait()
			return
		}
	}

}
