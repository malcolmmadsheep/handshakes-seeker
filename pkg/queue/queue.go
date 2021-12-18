package queue

import (
	"context"
	"time"
)

type Task = []byte

type ConsumerFunc = func(Task)

type Queue struct {
	tasks           chan Task
	config          Config
	stopConsumingCh chan struct{}
}

type Config struct {
	Delay     time.Duration
	QueueSize int
}

func New(config Config) Queue {
	return Queue{
		tasks:  make(chan Task, config.QueueSize),
		config: config,
	}
}

func (q *Queue) Publish(task Task) {
	q.tasks <- task
}

func (q *Queue) StopConsuming() {
	q.stopConsumingCh <- struct{}{}
}

func (q *Queue) StartConsuming(ctx context.Context) <-chan Task {
	consumeChan := make(chan Task)

	go func() {
		time.Sleep(q.config.Delay)

		select {
		case <-ctx.Done():
			return
		case <-q.stopConsumingCh:
			return
		case task := <-q.tasks:
			{
				consumeChan <- task
			}
		}
	}()

	return consumeChan
}
