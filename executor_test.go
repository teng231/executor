package executor

import (
	"context"
	"log"
	"testing"
	"time"
)

func testExecFn1(in ...interface{}) (interface{}, error) {
	log.Print("job ")
	time.Sleep(100 * time.Millisecond)
	return nil, nil
}
func testExecFn2(in ...interface{}) (interface{}, error) {
	log.Printf("job %v", in)
	time.Sleep(100 * time.Millisecond)
	return "responed", nil
}

func Test_executor(t *testing.T) {
	var engine IExecutor
	engine = CreateEngine(&EngineConfig{NumberWorker: 3, Capacity: 500})
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()
	engine.Run(ctx)
	testcases := make([]*Job, 0)
	testcases = append(testcases,
		&Job{
			Exectutor: testExecFn1,
		},
		&Job{
			Exectutor: testExecFn1,
		}, &Job{
			Params:    []interface{}{"param1", "param2"},
			Exectutor: testExecFn2,
			CallBack: func(result interface{}, err error) {
				log.Print("callback ", result, err)
			},
		})
	log.Print("job start send")
	for _, job := range testcases {
		if err := engine.Send(job); err != nil {
			log.Print(err)
		}
	}
	log.Print("job sent")
	time.AfterFunc(3*time.Second, func() {
		engine.Done()
	})
	engine.Wait()
}

// func Test_executorRescale(t *testing.T) {
// 	var engine IExecutor
// 	engine = CreateEngine(&EngineConfig{NumberWorker: 1, Capacity: 1000})
// 	ctx, cancelFn := context.WithCancel(context.Background())
// 	engine.Run(ctx)
// 	log.Print("job start send")
// 	for i := 0; i < 1000; i++ {
// 		engine.Send(&Job{
// 			Params:    []interface{}{"param1", i},
// 			Exectutor: testExecFn2,
// 		})
// 	}
// 	log.Print("job sent")
// 	time.Sleep(time.Second)
// 	cancelFn()

// 	ctx2, cancelFn2 := context.WithCancel(context.Background())
// 	defer cancelFn2()
// 	engine.Rescale(ctx2, 10)
// 	time.Sleep(10 * time.Second)
// }
