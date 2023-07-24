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

// func fnTest(i ...interface{}) (interface{}, error) {
// 	duration := i[0].(time.Duration)
// 	time.Sleep(duration * time.Millisecond)
// 	return nil, nil
// }

func fn100(i ...interface{}) (interface{}, error) {
	time.Sleep(100 * time.Millisecond)
	return "100", nil
}
func fn300(i ...interface{}) (interface{}, error) {
	time.Sleep(300 * time.Millisecond)
	return "300", nil
}
func TestRace2FnSimple(t *testing.T) {
	now := time.Now()
	out, err := Race2Task(fn300, fn100, 500*time.Millisecond)
	log.Print("run: ", time.Since(now))
	if err != nil {
		log.Print(err)
	}
	if out.(string) != "100" {
		log.Print(out)
	}

	now = time.Now()
	out, err = Race2Task(fn300, fn100, 50*time.Millisecond)
	log.Print("run: ", time.Since(now))
	if err != nil {
		log.Print(err)
	}
	log.Print(out, err)
	// if out.(string) != "100" {
	// 	log.Print(out)
	// }

}
