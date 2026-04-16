package worker

import (
	"context"
	"job-forge/internal/domain"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type JobHandler interface {
	Handle(ctx context.Context, job domain.Job) error
}

type JobMessage struct {
	Job      domain.Job
	Delivery amqp.Delivery
}

type Pool struct {
	workerCount int
	handler     JobHandler
}

func NewPool(workerCount int, handler JobHandler) *Pool {
	if workerCount <= 0 {
		workerCount = 1
	}
	return &Pool{
		workerCount: workerCount,
		handler:     handler,
	}
}

func (p *Pool) Start(ctx context.Context, jobs <-chan JobMessage) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(p.workerCount)

	for i := 0; i < p.workerCount; i++ {
		go func(workerID int) {
			defer wg.Done()

			for msg := range jobs {
				start := time.Now()
				err := p.handler.Handle(ctx, msg.Job)
				latency := time.Since(start)

				if err != nil {
					log.Printf("[worker %d] FAIL job_id=%s channel=%s latency=%s err=%v",
						workerID, msg.Job.ID, msg.Job.Channel, latency, err)

					_ = msg.Delivery.Nack(false, false)
					continue
				}

				log.Printf("[worker %d] OK job_id=%s channel=%s latency=%s",
					workerID, msg.Job.ID, msg.Job.Channel, latency)

				_ = msg.Delivery.Ack(false)
			}
		}(i)
	}

	return &wg
}
