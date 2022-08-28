package executor

import (
	"log"
	"sync"
	"testing"
	"time"
)

func testExec(in ...interface{}) (interface{}, error) {
	log.Print("job ", in)
	time.Sleep(3 * time.Second)
	return nil, nil
}
func TestSimpleSafeQueue(t *testing.T) {
	var engine ISafeQueue
	engine = CreateSafeQueue(&SafeQueueConfig{
		NumberWorkers: 3, Capacity: 500,
		WaitGroup: &sync.WaitGroup{},
	})
	defer engine.Close()
	engine.Run()
	testcases := make([]*Job, 0)
	testcases = append(testcases,
		&Job{
			Exectutor: testExec,
		},
		&Job{
			Exectutor: testExec,
		},
		&Job{
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
		Params:    []interface{}{"test 1"},
		Wg:        jwg,
	}
	engine.Send(job)
	job.Wait()
	for i := 0; i < 100; i++ {
		engine.Send(&Job{
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
		NumberWorkers: 2, Capacity: 500,
		WaitGroup: &sync.WaitGroup{},
	})
	defer engine.Close()
	engine.Run()
	group1 := make([]*Job, 0)
	group2 := make([]*Job, 0)
	now := time.Now()
	for i := 0; i < 10; i++ {
		group1 = append(group1, &Job{Exectutor: testExec, Params: []interface{}{"xxx", i}})
	}

	for i := 0; i < 20; i++ {
		group2 = append(group2, &Job{Exectutor: testExec, Params: []interface{}{"yyy", i}})
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
		group1 = append(group1, &Job{Exectutor: testExec, Params: []interface{}{"xxx", i}})
	}

	for i := 0; i < 20; i++ {
		group2 = append(group2, &Job{Exectutor: testExec, Params: []interface{}{"yyy", i}})
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
		engine.Send(&Job{Exectutor: testExec, Params: []interface{}{"xxx", i}})
	}
	log.Print("------------- done -------------- ", time.Since(now))
	engine.Send(&Job{Exectutor: testExec, Params: []interface{}{"over power"}})
	log.Print("------------- done2 -------------- ", time.Since(now))
	<-wait
}
