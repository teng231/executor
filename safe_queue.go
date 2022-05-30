package executor

import (
	"errors"
	"log"
	"sync"
	"time"
)

type FnWithTerminating = func() []*Job

type Job struct {
	Params    []interface{}
	Exectutor func(...interface{}) (interface{}, error)
	CallBack  func(interface{}, error)
	Wg        *sync.WaitGroup
}

func (j *Job) Wait() {
	if j.Wg == nil {
		return
	}
	j.Wg.Add(1)
	j.Wg.Wait()
}

func (j *Job) Done() {
	if j.Wg == nil {
		return
	}
	j.Wg.Done()
}

type SafeQueueConfig struct {
	NumberWorkers int
	Capacity      int
	WaitGroup     *sync.WaitGroup
}

type SafeQueueInfo struct {
	LenJobs       int
	NumberWorkers int
	LenLocks      int
}
type ISafeQueue interface {
	Info() SafeQueueInfo
	Close() error
	RescaleUp(numWorker uint)
	RescaleDown(numWorker uint) error
	Run()
	Send(jobs ...*Job) error
	Wait()
	Done()
}
type SafeQueue struct {
	hub                chan *Job
	numberWorkers      int
	closech            map[string](chan bool)
	wg                 *sync.WaitGroup
	TerminatingHandler FnWithTerminating
}

func CreateSafeQueue(config *SafeQueueConfig) *SafeQueue {
	if config.Capacity == 0 {
		config.Capacity = default_capacity_hub
	}
	if config.NumberWorkers == 0 {
		config.NumberWorkers = default_numberworker
	}
	if config.WaitGroup == nil {
		config.WaitGroup = &sync.WaitGroup{}
	}
	engine := &SafeQueue{
		hub:           make(chan *Job, config.Capacity),
		numberWorkers: config.NumberWorkers,
		closech:       make(map[string]chan bool),
		wg:            &sync.WaitGroup{},
	}
	return engine
}

func (s *SafeQueue) Info() SafeQueueInfo {
	return SafeQueueInfo{
		LenJobs:       len(s.hub),
		NumberWorkers: s.numberWorkers,
		LenLocks:      len(s.closech),
	}
}
func (s *SafeQueue) Wait() {
	s.wg.Add(1)
	s.wg.Wait()
}

func (s *SafeQueue) Done() {
	s.wg.Done()
}

func (s *SafeQueue) Close() error {
	close(s.hub)
	for _, ch := range s.closech {
		close(ch)
	}
	return nil
}
func (s *SafeQueue) Send(jobs ...*Job) error {
	for _, j := range jobs {
		if j.Exectutor == nil {
			return errors.New(E_exector_required)
		}
	}
	for _, j := range jobs {
		s.hub <- j
	}
	return nil
}

// RescaleUp numberWorker limit at 10
func (s *SafeQueue) RescaleUp(numWorker uint) {
	if numWorker > 10 {
		numWorker = 10
	}
	s.workerStart(int(numWorker))
	s.numberWorkers += int(numWorker)
}

func (s *SafeQueue) RescaleDown(numWorker uint) error {
	if int(numWorker) >= s.numberWorkers {
		return errors.New("numberworker scale down over total worker")
	}
	i := 0
	for key, ch := range s.closech {
		log.Print("close ", key)
		if i == int(numWorker)-1 {
			break
		}
		i++
		ch <- true
	}
	s.numberWorkers -= int(numWorker)
	return nil
}

func (s *SafeQueue) Run() {
	s.workerStart(s.numberWorkers)
}

func (s *SafeQueue) Jobs() []*Job {
	jobs := make([]*Job, 0, 5)
	for job := range s.hub {
		jobs = append(jobs, job)
	}
	return jobs
}

func (s *SafeQueue) workerStart(worker int) {
	for i := 0; i < worker; i++ {
		key := RandStringRunes(10)
		s.closech[key] = make(chan bool)
		go func(ch chan bool, key string) {
			for {
				select {
				case job := <-s.hub:
					result, err := job.Exectutor(job.Params...)
					if job.CallBack != nil {
						job.CallBack(result, err)
					}
					time.Sleep(3 * time.Second)
					job.Done()
				case <-ch:
					log.Print("end of worker id ", key)
					return
				}
			}
		}(s.closech[key], key)
	}
}