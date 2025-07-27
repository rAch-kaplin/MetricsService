package workerpool

import (
	"context"
	"sync"
)

type Task struct {
	Execute func()
}


type WorkerPool struct {
	workers int
	tasks   chan Task
	wg      sync.WaitGroup
}

func New(workers int) *WorkerPool {
	return &WorkerPool{
		workers: workers,
		tasks:   make(chan Task, workers),
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go worker(ctx, &wp.wg, wp.tasks)
	}
}

func (wp *WorkerPool) AddTask(task *Task) {
	wp.tasks <- *task
}

func (wp *WorkerPool) Wait() {
	close(wp.tasks)
	wp.wg.Wait()
}

func worker(ctx context.Context, wg *sync.WaitGroup, tasks <- chan Task) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		case task, ok := <-tasks:
			if !ok {
				return
			}
			task.Execute()
		}
	}
}
