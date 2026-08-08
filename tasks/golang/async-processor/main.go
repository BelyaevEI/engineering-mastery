package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type Task interface {
	Do() (int, error)
}

type Tasker struct {
}

func (t *Tasker) Do() (int, error) {
	items := []int{1, 2, 3, 4, 5}
	now := time.Now()
	if now.Second()%2 == 0 {
		return items[rand.Intn(len(items))], nil
	}

	return 0, errors.New("time is too small")
}

type Processor struct {
	taskChan      chan Task
	resultChan    chan int
	errorChan     chan error
	wg            sync.WaitGroup
	closed        bool
	mu            sync.RWMutex
	ctx           context.Context
	cancelFunc    context.CancelFunc
	workerCount   int
	completedTask int32
}

func NewProcessor(workerCount int, queue int) (*Processor, error) {
	if workerCount < 0 {
		return nil, fmt.Errorf("invalid worker count: %d", workerCount)
	}

	if queue < 0 {
		return nil, fmt.Errorf("invalid queue size: %d", queue)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Processor{
		taskChan:    make(chan Task, queue),
		resultChan:  make(chan int, queue),
		errorChan:   make(chan error, queue),
		workerCount: workerCount,
		ctx:         ctx,
		cancelFunc:  cancel,
	}, nil
}

func (p *Processor) Start() {
	p.wg.Add(p.workerCount)
	for i := 0; i < p.workerCount; i++ {

		go func() {
			defer p.wg.Done()
			for {
				select {
				case task, ok := <-p.taskChan:
					if !ok {
						return
					}
					res, err := task.Do()
					if err != nil {
						select {
						case p.errorChan <- err:
						default:
							fmt.Println("lock")
						}
					}
					select {
					case p.resultChan <- res:
					default:
						fmt.Println("lock")
					}
					atomic.AddInt32(&p.completedTask, 1)
					fmt.Printf("task completed: %d\n", atomic.LoadInt32(&p.completedTask))
				case <-p.ctx.Done():
					return
				}
			}
		}()

	}
}

func (p *Processor) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	close(p.taskChan)
	p.closed = true
}

func (p *Processor) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return
	}
	close(p.taskChan)
	p.wg.Wait()
	p.closed = true
	p.cancelFunc()
}

func (p *Processor) Push(task Task) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if !p.closed {
		select {
		case p.taskChan <- task:
		default:
			return fmt.Errorf("task queue is full")
		}
	}

	return nil
}

func main() {

	p, err := NewProcessor(10, 50)
	if err != nil {
		fmt.Println(err)
	}

	go p.Start()

	for i := 0; i < 10; i++ {
		t := Tasker{}
		p.Push(&t)
	}

	go func() {
		for res := range p.resultChan {
			fmt.Printf("result: %d\n", res)
		}
	}()

	go func() {
		for err := range p.errorChan {
			fmt.Println(err)
		}
	}()

	time.Sleep(1 * time.Second)
	p.Close()
}
