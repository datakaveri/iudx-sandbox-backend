package taskqueue

import (
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/vmihailenco/taskq/v3"
	"github.com/vmihailenco/taskq/v3/memqueue"
)

type TaskQueue struct {
	Queue taskq.Queue
}

func InitializeTaskQueue() *TaskQueue {
	// Create a queue factory.
	var QueueFactory = memqueue.NewFactory()

	// Create a queue.
	var MainQueue = QueueFactory.RegisterQueue(&taskq.QueueOptions{
		Name:    "sandbox-worker",
		Storage: taskq.NewLocalStorage(),
	})

	if err := MainQueue.Consumer().Stop(); err != nil {
		logger.Error.Println(err.Error())
	}

	return &TaskQueue{
		Queue: MainQueue,
	}
}
