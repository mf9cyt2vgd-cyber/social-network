package main

import (
	"context"
	"fmt"
	"notifications-service/internal/config"
	"notifications-service/internal/consumer"
	"notifications-service/internal/lib/logger"
	"notifications-service/internal/lib/utils"
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

	c, err := consumer.NewKafkaConsumer(cfg.Brokers, cfg.GroupID, cfg.Topic, log)
	if err != nil {
		log.Error("failed to create consumer", "error", err)
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
			post, err := utils.ConvertKafkaMessageIntoPost(msg)
			if err != nil {
				log.Error("failed to convert message", "error", err)
				continue
			}
			log.Info("successfully read post from Kafka", "post.id", post.ID)
			fmt.Println(post)
		case e := <-errChan:
			log.Error("got error from consumer", "error", e)
			continue
		case <-done:
			log.Info("got interrupt signal")
			cancel()
			wg.Wait()
			return
		}
	}

}
