package executor

import (
	"errors"
	"log"
	"sync"
)

type FnWithTerminating = func() []*Job

type Job struct {
	Params    []interface{}
	Exectutor func(...interface{}) (interface{}, error)
	CallBack  func(interface{}, error)
	Wg        *sync.WaitGroup
	Id        interface{}
}

var CallBacks func([]interface{}, error)

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
	Info() SafeQueueInfo              // engine info
	Close() error                     // close all anything
	RescaleUp(numWorker uint)         // increase worker
	RescaleDown(numWorker uint) error // reduce worker
	Run()                             // start
	Send(jobs ...*Job) error          // push job to hub
	SendWithGroup(jobs ...*Job) error // push job to hub and wait to done
	Wait()                            // keep block thread
	Done()                            // Immediate stop wait
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

func RunSafeQueue(config *SafeQueueConfig) *SafeQueue {
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
	engine.Run()
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
	if s.wg == nil {
		return
	}
	s.wg.Add(1)
	s.wg.Wait()
}

func (s *SafeQueue) Done() {
	if s.wg == nil {
		return
	}
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

func (s *SafeQueue) SendWithGroup(jobs ...*Job) error {
	if len(jobs) == 0 {
		return nil
	}
	wg := &sync.WaitGroup{}
	wg.Add(len(jobs))
	for _, j := range jobs {
		j.Wg = wg
		s.hub <- j
	}
	wg.Wait()
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
	handleFn := func(ch chan bool, key string) {
		for {
			select {
			case job := <-s.hub:
				result, err := job.Exectutor(job.Params...)
				if job.CallBack != nil {
					job.CallBack(result, err)
				}
				if job.Wg != nil {
					job.Done()
				}
			case <-ch:
				log.Print("end of worker id ", key)
				return
			}
		}
	}
	for i := 0; i < worker; i++ {
		key := RandStringRunes(10)
		s.closech[key] = make(chan bool)
		go handleFn(s.closech[key], key)
	}
}
