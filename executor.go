package executor

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	default_capacity_hub = 2000
	default_numberworker = 3
	E_invalid_priority   = "invalid priority"
	E_exector_required   = "exector required"
)

type Job struct {
	Params    []interface{}
	Exectutor func(...interface{}) (interface{}, error)
	CallBack  func(interface{}, error)
}

type Engine struct {
	hub           chan *Job
	numberWorkers int
	wg            *sync.WaitGroup
	ctx           context.Context
}

type IExecutor interface {
	Run(context.Context, func([]*Job) error)
	Send(*Job) error
	Rescale(context.Context, int) error
	Len() int
	GetHubs() []*Job
	Wait()
	Done()
}

type EngineConfig struct {
	NumberWorker int
	Capacity     int
	WaitGroup    *sync.WaitGroup
}

func CreateEngine(config *EngineConfig) *Engine {
	if config.Capacity == 0 {
		config.Capacity = default_capacity_hub
	}
	if config.NumberWorker == 0 {
		config.NumberWorker = default_numberworker
	}
	if config.WaitGroup == nil {
		config.WaitGroup = &sync.WaitGroup{}
	}
	config.WaitGroup.Add(1)
	engine := &Engine{
		hub:           make(chan *Job, config.Capacity),
		numberWorkers: config.NumberWorker,
		wg:            config.WaitGroup,
	}
	return engine
}

func (e *Engine) Rescale(ctx context.Context, numberWorker int) error {
	if numberWorker == 0 {
		numberWorker = default_numberworker
	}
	e.numberWorkers = numberWorker
	e.Run(ctx, nil)
	return nil
}

func (e *Engine) Len() int {
	return len(e.hub)
}

func (e *Engine) Wait() {
	e.wg.Wait()
}

func (e *Engine) Done() {
	close(e.hub)
	e.wg.Done()
}

func (e *Engine) Send(j *Job) error {
	if j.Exectutor == nil {
		return errors.New(E_exector_required)
	}
	e.hub <- j
	return nil
}
func (e *Engine) GetHubs() []*Job {
	hubs := make([]*Job, 0, 10)
	for item := range e.hub {
		hubs = append(hubs, item)
	}
	return hubs
}

func (e *Engine) Run(ctx context.Context, terminateHandle func([]*Job) error) {
	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	e.ctx = ctx
	for i := 0; i < e.numberWorkers; i++ {
		go func() {
			for {
				select {
				case job := <-e.hub:
					e.wg.Add(1)
					result, err := job.Exectutor(job.Params...)
					if job.CallBack != nil {
						job.CallBack(result, err)
					}
					e.wg.Done()
				case <-signChan:
					log.Print("terminating processing ", e.hub)
					close(e.hub)
					if len(e.hub) == 0 {
						return
					}

					if terminateHandle != nil {
						terminateHandle(e.GetHubs())
						return
					}

					for job := range e.hub {
						e.wg.Add(1)
						go job.Exectutor(job.Params...)
						job.CallBack = func(i interface{}, _ error) {
							e.wg.Done()
						}
					}
					log.Print("terminated")
				case <-ctx.Done():
					log.Print("done")
					return
				}
			}
		}()
	}
}
