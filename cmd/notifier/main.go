package main

import (
	"context"
	"job-forge/internal/channels"
	"job-forge/internal/channels/email"
	"job-forge/internal/config"
	"job-forge/internal/domain"
	"job-forge/internal/queue/rabbitmq"
	"job-forge/internal/worker"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conn, err := rabbitmq.New(
		rootCtx,
		cfg.RabbitMQURL,
		cfg.RabbitMQQueue,
		cfg.RabbitMQExchange,
		cfg.RabbitMQRoutingKey,
	)

	if err != nil {
		log.Fatalf("rabbitmq connect: %v", err)
	}

	defer func() { _ = conn.Close() }()

	ch := conn.Channel()

	if err := ch.Qos(cfg.WorkerCount, 0, false); err != nil {
		log.Fatalf("rabbitmq qos: %v", err)
	}

	deliveries, err := ch.Consume(
		cfg.RabbitMQQueue,
		"",
		false, // autoAck = false
		false, // exclusive
		false, // noLocal (não existe no AMQP; lib ignora)
		false, // noWait
		nil,   // args
	)

	if err != nil {
		log.Fatalf("rabbitmq consume: %v", err)
	}

	router := channels.NewRouter(map[string]channels.Handler{
		domain.ChannelEmail: email.NewHandler(),
	})

	jobsCh := make(chan worker.JobMessage, cfg.WorkerCount*2)
	pool := worker.NewPool(cfg.WorkerCount, router)
	workerWG := pool.Start(rootCtx, jobsCh)

	var consumeWG sync.WaitGroup
	consumeWG.Add(1)

	go func() {
		defer consumeWG.Done()
		defer close(jobsCh)

		for {
			select {
			case <-rootCtx.Done():
				return
			case d, ok := <-deliveries:
				if !ok {
					return
				}

				job, err := domain.DecodeJob(d.Body)
				if err != nil {
					log.Printf("[consumer] invalid job: %v", err)
					_ = d.Nack(false, false)
					continue
				}

				select {
				case jobsCh <- worker.JobMessage{Job: job, Delivery: d}:
					//
				case <-rootCtx.Done():
					_ = d.Nack(false, false)
					return
				}
			}
		}
	}()

	<-rootCtx.Done()
	log.Printf("[main] shutdown requested")

	consumeWG.Wait()
	workerWG.Wait()

	log.Printf("[main] shutdown complete")
	_ = amqp.Publishing{}
}
