package executor

import (
	"log"
	"sync"
	"testing"
	"time"
)

func TestRaceSimple(t *testing.T) {
	now := time.Now()
	_, err := Race(200*time.Millisecond, func(i ...interface{}) (interface{}, error) {
		time.Sleep(time.Duration(i[0].(int)) * time.Millisecond)
		return nil, nil
	}, 30)
	log.Print("run: ", time.Since(now))
	if err != nil {
		t.Fail()
	}
	now = time.Now()
	_, err = Race(200*time.Millisecond, func(i ...interface{}) (interface{}, error) {
		time.Sleep(time.Duration(i[0].(int)) * time.Millisecond)
		return nil, nil
	}, 500)
	log.Print("run: ", time.Since(now))
	if err == nil {
		t.Fail()
	}
	log.Print(err)
	if err.Error() != E_TimeUp {
		t.Fail()
	}
}

func TestRaceWithExecutor(t *testing.T) {
	engine := RunSafeQueue(&SafeQueueConfig{
		NumberWorkers: 3, Capacity: 500,
		WaitGroup: &sync.WaitGroup{},
	})
	now := time.Now()
	_, err := RaceWithSafeQueue(engine, 200*time.Millisecond, func(i ...interface{}) (interface{}, error) {
		time.Sleep(time.Duration(i[0].(int)) * time.Millisecond)
		return nil, nil
	}, 30)
	log.Print("run: ", time.Since(now))
	if err != nil {
		t.Fail()
	}
	now = time.Now()
	_, err = RaceWithSafeQueue(engine, 200*time.Millisecond, func(i ...interface{}) (interface{}, error) {
		time.Sleep(time.Duration(i[0].(int)) * time.Millisecond)
		return nil, nil
	}, 500)
	log.Print("run: ", time.Since(now))
	if err == nil {
		t.Fail()
	}
	log.Print(err)
	if err.Error() != E_TimeUp {
		t.Fail()
	}
}
