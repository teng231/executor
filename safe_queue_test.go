package executor

import (
	"errors"
	"log"
	"strconv"
	"sync"
	"testing"
	"time"
)

func testExec(in ...interface{}) (interface{}, error) {
	log.Print("job ", in)
	t := in[0].(int)
	text := ""
	if in[1] != nil {
		text = in[1].(string)
	}
	time.Sleep(time.Duration(t) * time.Second)
	if text == "fail" {
		return nil, errors.New("fail")
	}
	return nil, nil
}
func TestSimpleSafeQueue(t *testing.T) {
	var engine ISafeQueue
	engine = RunSafeQueue(&SafeQueueConfig{
		NumberWorkers: 3, Capacity: 500,
		WaitGroup: &sync.WaitGroup{},
	})
	defer engine.Close()
	engine.Run()
	testcases := make([]*Job, 0)
	testcases = append(testcases,
		&Job{
			Params:    []interface{}{1, "xxx"},
			Exectutor: testExec,
		},
		&Job{
			Params:    []interface{}{1, "xxx"},
			Exectutor: testExec,
		},
		&Job{Params: []interface{}{1, "xxx"},
			Exectutor: testExec,
			// Wg:        &sync.WaitGroup{},
		},
	)
	log.Print("job start send")
	engine.Send(testcases...)
	log.Print("------------------- test wait for callback -------------------")
	jwg := &sync.WaitGroup{}
	job := &Job{
		Exectutor: testExec,
		Params:    []interface{}{1, "test 1"},
		Wg:        jwg,
	}
	engine.Send(job)
	job.Wait()
	for i := 0; i < 100; i++ {
		engine.Send(&Job{
			Params:    []interface{}{1, "xx"},
			Exectutor: testExec,
		})
	}
	log.Print("------------------- test scale up size -------------------")
	log.Print(engine.Info())

	engine.RescaleUp(5)
	log.Print(engine.Info())

	log.Print("------------------- test scaledown size -------------------")
	log.Print(engine.Info())

	engine.RescaleDown(3)
	log.Print("aaaaa", engine.Info())

	time.AfterFunc(5*time.Second, func() {
		log.Print("xxxxxxx: ", engine.Info())
	})
	engine.Wait()
}

func TestSimpleSafeQueueGroup(t *testing.T) {
	var engine ISafeQueue
	engine = CreateSafeQueue(&SafeQueueConfig{
		NumberWorkers: 8, Capacity: 500,
		WaitGroup: &sync.WaitGroup{},
	})
	// defer engine.Close()
	engine.Run()
	group1 := make([]*Job, 0)
	group2 := make([]*Job, 0)
	now := time.Now()
	for i := 0; i < 5; i++ {
		group1 = append(group1, &Job{Exectutor: testExec, Params: []interface{}{i, "xxx", i}})
	}

	for i := 0; i < 4; i++ {
		group2 = append(group2, &Job{Exectutor: testExec, Params: []interface{}{i, "yyy", i}})
	}
	engine.SendWithGroup(group1...)
	engine.SendWithGroup(group2...)
	log.Print("------------- done -------------- ", time.Since(now))
	engine.Wait()
}

func TestNotUsingWaitgroup(t *testing.T) {
	var engine ISafeQueue
	engine = CreateSafeQueue(&SafeQueueConfig{
		NumberWorkers: 2, Capacity: 500,
		// WaitGroup: &sync.WaitGroup{},
	})
	wait := make(chan bool)
	defer engine.Close()
	engine.Run()
	group1 := make([]*Job, 0)
	group2 := make([]*Job, 0)
	now := time.Now()
	for i := 0; i < 10; i++ {
		group1 = append(group1, &Job{Exectutor: testExec, Params: []interface{}{i, "xxx", i}})
	}

	for i := 0; i < 20; i++ {
		group2 = append(group2, &Job{Exectutor: testExec, Params: []interface{}{i, "yyy", i}})
	}
	engine.SendWithGroup(group1...)
	engine.SendWithGroup(group2...)
	log.Print("------------- done -------------- ", time.Since(now))
	// engine.Wait()
	<-wait
}

func TestOverload(t *testing.T) {
	var engine ISafeQueue
	engine = CreateSafeQueue(&SafeQueueConfig{
		NumberWorkers: 2, Capacity: 30,
	})
	wait := make(chan bool)
	defer engine.Close()
	engine.Run()
	now := time.Now()
	for i := 0; i < 30; i++ {
		engine.Send(&Job{Exectutor: testExec, Params: []interface{}{i, "xxx", i}})
	}
	log.Print("------------- done -------------- ", time.Since(now))
	engine.Send(&Job{Exectutor: testExec, Params: []interface{}{"over power"}})
	log.Print("------------- done2 -------------- ", time.Since(now))
	<-wait
}
func TestGroup(t *testing.T) {
	var engine ISafeQueue
	engine = CreateSafeQueue(&SafeQueueConfig{
		NumberWorkers: 3, Capacity: 30,
	})
	wait := make(chan bool)
	// defer engine.Close()
	engine.Run()
	now := time.Now()
	for i := 0; i < 10; i++ {
		if i == 3 {
			engine.SendWithGroup(&Job{Exectutor: testExec, Params: []interface{}{i, "fail"},
				CallBack: func(i interface{}, err error) {
					log.Print(i, err)
				}})
			continue
		}
		engine.SendWithGroup(&Job{Exectutor: testExec, Params: []interface{}{i, "xxxx"},
			CallBack: func(i interface{}, err error) {
				log.Print(i, err)
			}})
	}
	log.Print("------------- done2 -------------- ", time.Since(now))
	<-wait
}

func TestShowInfo(t *testing.T) {
	var engine ISafeQueue
	engine = RunSafeQueue(&SafeQueueConfig{
		NumberWorkers: 3, Capacity: 50,
	})
	wait := make(chan bool)
	// defer engine.Close()
	now := time.Now()
	// go func() {
	// 	for {
	// 		time.Sleep(100 * time.Millisecond)
	// 		log.Print(engine.Info())
	// 	}
	// }()
	for i := 0; i < 100; i++ {
		engine.Send(&Job{
			Params: []interface{}{i},
			Exectutor: func(i ...interface{}) (interface{}, error) {
				time.Sleep(200 * time.Millisecond)
				log.Print("i:", i)
				return nil, nil
			},
		})
	}
	log.Print("------------- done2 -------------- ", time.Since(now))
	<-wait
}

func TestRunQueue(t *testing.T) {
	var engine ISafeQueue
	engine = RunSafeQueue(&SafeQueueConfig{
		NumberWorkers: 3, Capacity: 200,
	})
	wait := make(chan bool)
	// defer engine.Close()
	now := time.Now()
	// go func() {
	// 	for {
	// 		time.Sleep(100 * time.Millisecond)
	// 		log.Print(engine.Info())
	// 	}
	// }()
	for i := 0; i < 100; i++ {
		engine.Send(&Job{
			Params: []interface{}{i},
			Exectutor: func(i ...interface{}) (interface{}, error) {
				time.Sleep(100 * time.Millisecond)
				log.Print("i:", i)
				return nil, nil
			},
		})
	}
	log.Print("------------- done2 -------------- ", time.Since(now))
	<-wait
}

func TestRunWithLimiter(t *testing.T) {
	var engine ISafeQueue

	engine = RunSafeQueue(&SafeQueueConfig{
		NumberWorkers: 100, Capacity: 200,
	})
	wait := make(chan bool)
	// defer engine.Close()
	now := time.Now()
	limiter := &Limiter{Key: "any1", Limit: 2}
	for i := 0; i < 100; i++ {
		engine.Send(&Job{
			Params: []interface{}{i},
			Exectutor: func(i ...interface{}) (interface{}, error) {
				time.Sleep(100 * time.Millisecond)
				log.Print("i:", i)
				return nil, nil
			},
			Limiter: limiter,
		})
		engine.Send(&Job{
			Params: []interface{}{i},
			Exectutor: func(i ...interface{}) (interface{}, error) {
				time.Sleep(100 * time.Millisecond)
				log.Print("i2:", i)
				return nil, nil
			},
		})
	}
	log.Print("------------- done2 -------------- ", time.Since(now))
	<-wait
}

func TestRunQueueErrorRetry(t *testing.T) {
	var engine ISafeQueue
	engine = RunSafeQueue(&SafeQueueConfig{
		NumberWorkers: 3, Capacity: 200,
	})
	done := map[string]bool{}
	mt := &sync.Mutex{}
	wait := make(chan bool)
	f := func(i ...interface{}) (interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		log.Print("i:", i)
		val := i[0].(int)
		if val == 5 {
			return val, errors.New("st")
		}
		mt.Lock()
		done["hello_"+strconv.Itoa(val)] = true
		mt.Unlock()
		return nil, nil
	}
	jobs := []*Job{}
	backupJobs := []*Job{}
	for i := 0; i < 10; i++ {
		jobs = append(jobs, &Job{
			Params:    []interface{}{i},
			Exectutor: f,
			CallBack: func(i interface{}, err error) {
				if err != nil {
					log.Print("eeeee ", i)
					time.Sleep(1000 * time.Millisecond)
					backupJobs = append(backupJobs, &Job{
						Params:    []interface{}{i.(int) + 10},
						Exectutor: f,
					})
				}
			},
		})
	}
	engine.SendWithGroup(jobs...)
	if len(backupJobs) > 0 {
		engine.SendWithGroup(backupJobs...)
	}
	log.Print(done)
	<-wait
}
