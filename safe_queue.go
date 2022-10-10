package executor

import (
	"errors"
	"log"
	"sync"
	"time"
)

const (
	CANCELLED = "cancelled"
)
const (
	ON_WORKING = 1 + iota
	ON_FAILURE
)

type FnWithTerminating = func() []*Job

type Job struct {
	Params                     []interface{}
	Exectutor                  func(...interface{}) (interface{}, error)
	CallBack                   func(interface{}, error)
	Wg                         *sync.WaitGroup
	GroupId                    string
	IsCancelWhenSomethingError bool
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
	Debug         bool // enable with show log and show info after 5 min
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
	MakeGroupIdAndStartQueue() string
	ReleaseGroupId(groupId string)
	Wait() // keep block thread
	Done() // Immediate stop wait
}
type SafeQueue struct {
	hub                chan *Job
	numberWorkers      int
	closech            map[string](chan bool)
	wg                 *sync.WaitGroup
	TerminatingHandler FnWithTerminating
	mGroupIdStatus     map[string]int
	mtlock             *sync.RWMutex
}

func (s *SafeQueue) MakeGroupIdAndStartQueue() string {
	groupId := RandStringRunes(12)
	s.mtlock.Lock()
	s.mGroupIdStatus[groupId] = ON_WORKING
	s.mtlock.Unlock()
	return groupId
}

func (s *SafeQueue) ReleaseGroupId(groupId string) {
	s.mtlock.Lock()
	delete(s.mGroupIdStatus, groupId)
	s.mtlock.Unlock()
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
		hub:            make(chan *Job, config.Capacity),
		numberWorkers:  config.NumberWorkers,
		closech:        make(map[string]chan bool),
		wg:             &sync.WaitGroup{},
		mGroupIdStatus: make(map[string]int),
		mtlock:         &sync.RWMutex{},
	}
	if config.Debug {
		tick := time.NewTicker(5 * time.Minute)
		go func() {
			for {
				<-tick.C
				info := engine.Info()
				log.Print("[SAFE_QUEUE_DEBUG]: ", info)
			}
		}()
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
		hub:            make(chan *Job, config.Capacity),
		numberWorkers:  config.NumberWorkers,
		closech:        make(map[string]chan bool),
		wg:             &sync.WaitGroup{},
		mGroupIdStatus: make(map[string]int),
		mtlock:         &sync.RWMutex{},
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

				s.mtlock.RLock()
				status := s.mGroupIdStatus[job.GroupId]
				s.mtlock.RUnlock()

				if status == ON_FAILURE && job.IsCancelWhenSomethingError {
					if job.CallBack != nil {
						job.CallBack(nil, errors.New(CANCELLED))
					}
					job.Done()
					continue
				}
				result, err := job.Exectutor(job.Params...)
				if err != nil {
					s.mtlock.Lock()
					s.mGroupIdStatus[job.GroupId] = ON_FAILURE
					s.mtlock.Unlock()
				}
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
