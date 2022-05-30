package executor

import (
	"log"
	"sync"
	"testing"
	"time"
)

func testExec(in ...interface{}) (interface{}, error) {
	log.Print("job ")
	time.Sleep(50 * time.Millisecond)
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
		NumberWorkers: 3, Capacity: 500,
		WaitGroup: &sync.WaitGroup{},
	})
	defer engine.Close()
	engine.Run()
	group1 := make([]*Job, 0)
	now := time.Now()
	for i := 0; i < 10; i++ {
		group1 = append(group1, &Job{Exectutor: testExec})
	}
	engine.SendWithGroup(group1...)
	log.Print("------------- done -------------- ", time.Since(now))
	engine.Wait()
}
