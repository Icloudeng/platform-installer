package queue

import (
	"context"

	"github.com/golang-queue/queue"
)

var Queue *queue.Queue

func init() {
	// Proccess only one queue
	Queue = queue.NewPool(1)

	Queue.QueueTask(func(ctx context.Context) error {
		return nil
	})
}
