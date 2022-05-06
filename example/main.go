package main

import (
	"context"
	"log"
	"time"

	executor "github.com/teng231/executor"
)

func main() {
	terminateBehaves := make(map[executor.Priority]func([]*executor.Job))
	terminateBehaves[executor.Priority_common] = func([]*executor.Job) {

	}
	config := &executor.EngineConfig{
		NumberWorker:    1,
		Capacity:        300,
		TerminateBehave: terminateBehaves,
	}
	engine := executor.CreateEngine(config)
	engine.Run(context.Background())
	for i := 0; i < 300; i++ {
		engine.Send(&executor.Job{
			Params:   []interface{}{i},
			Priority: executor.Priority_common,
			Exectutor: func(i ...interface{}) (interface{}, error) {
				log.Print(" run ", i)
				time.Sleep(100 * time.Millisecond)
				return i, nil
			},
		})
	}
	log.Print("sent")
}
