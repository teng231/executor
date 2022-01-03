package executor

import (
	"context"
	"errors"
	"log"
)

type Priority int

const (
	Priority_immediate   = 1
	Priority_common      = 2
	default_capacity_hub = 2000
	default_numberworker = 3
	E_invalid_priority   = "invalid priority"
	E_exector_required   = "exector required"
)

type Job struct {
	Params    []interface{}
	Exectutor func(...interface{}) (interface{}, error)
	CallBack  func(interface{}, error)
	Priority  Priority
}

type Engine struct {
	immediateHub  chan *Job
	commonHub     chan *Job
	numberWorkers int
	capacity      int
}

type IExecutor interface {
	Run(context.Context)
	Send(*Job) error
	Rescale(context.Context, int) error
}

func CreateEngine(numberWorker, capacity int) *Engine {
	if capacity == 0 {
		capacity = default_capacity_hub
	}
	if numberWorker == 0 {
		numberWorker = default_numberworker
	}
	return &Engine{
		immediateHub:  make(chan *Job, capacity),
		commonHub:     make(chan *Job, capacity),
		numberWorkers: numberWorker,
		capacity:      capacity,
	}
}
func (e *Engine) Rescale(ctx context.Context, numberWorker int) error {
	if numberWorker == 0 {
		numberWorker = default_numberworker
	}
	e.numberWorkers = numberWorker
	e.Run(ctx)
	return nil
}

func (e *Engine) Send(j *Job) error {
	if j.Priority == 0 || (j.Priority != Priority_common && j.Priority != Priority_immediate) {
		return errors.New(E_invalid_priority)
	}
	if j.Exectutor == nil {
		return errors.New(E_exector_required)
	}
	if j.Priority == Priority_immediate {
		e.immediateHub <- j
	}
	if j.Priority == Priority_common {
		e.commonHub <- j
	}
	return nil
}

func (e *Engine) Run(ctx context.Context) {
	for i := 0; i < e.numberWorkers; i++ {
		go func() {
			for {
				select {
				case job := <-e.immediateHub:
					result, err := job.Exectutor(job.Params...)
					if job.CallBack != nil {
						job.CallBack(result, err)
					}
				case <-ctx.Done():
					log.Print("done")
					return
				}
			}
		}()
	}
	for i := 0; i < e.numberWorkers; i++ {
		go func() {
			for {
				select {
				case job := <-e.commonHub:
					result, err := job.Exectutor(job.Params...)
					if job.CallBack != nil {
						job.CallBack(result, err)
					}
				case <-ctx.Done():
					log.Print("done")
					return
				}
			}
		}()
	}
}
