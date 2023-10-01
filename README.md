# executor
fast exec task with go and less mem ops


## Why we need executor?

Go with goroutine is the best structure, it's help developer folk a thread easy than other language.
Go have chanel mechanic to connect goroutine(thread) data.
But what problem??

- First, `go` keyword folk goroutine but go engine not commit execute duration time taget. Then you want exact your goroutine can run but you don't know when? It's ok fast with small traffic but when server high load you will scare about that, because is slow, and when you create many goroutine is very slow. `waste of time`

- Second, when you create and release goroutine many time, it's with increase memory and CPU time of system, Sometime some developer forgot shut it down, It's can be make app heavy because leak memory. `waste of resource`

## How to resolve problem??

This's how i resolved this the problem.

- We create `exact` or `dynamic` number goroutines. Then using `Job` is a unit bring data information to `Worker` to process. Worker no need release and you only create 1 time or reset when you update config.

- Job bring 2 importance field is: `params` and `executor` and you can call them like `executor(params)` and get all result and return on `callback` is optional
- It's help your app no need create and release goroutine continuity.
- Easy to access and using when coding.

## Disadvance?

- At you know. If you using callback maybe memory leak if you not control bester.
- If buffer chanel full you need wait a little bits for execute next job.

## Installtion

Now it's time you can install lib and experience

```bash
go get github.com/teng231/executor
```

## Usage : Interface list
```go
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
	Jobs() []*Job
	TerminatingHandler() []*Job
}
```

### Initial

```go
    engine := RunSafeQueue(&SafeQueueConfig{
        NumberWorkers: 3,
        Capacity: 500,
    })
```

### Send Simple Job and executed
```go
    // limmiter it a ratelimit limit 10 execs/sec
    limiter := &Limiter{Key: "Something", Limit: 10}
    // simple job
    j := &Job{
        Limiter:limiter,
        Params: []interface{1, "abc"},
        Exectutor: func(in ...interface{}) {
            // any thing
            log.Print("do it")
        },
    }
    engine.Send(j)
    // send mutiple job
    jobs := []*Job{
        {
        Exectutor: func(in ...interface{}) {
            // any thing
        },
        Params: []interface{1, "abc"}
        },
         Exectutor: func(in ...interface{}) {
            // any thing
        },
        Params: []interface{2, "abc"}
    }
    engine.Send(jobs...)
```

### Send Job with groups
```go
    limiter := &Limiter{Key: "Something", Limit: 10}
    // prepaire a group job.
	group1 := make([]*Job, 0)
	for i := 0; i < 10; i++ {
		group1 = append(group1, &Job{
            Limiter:limiter,
            Exectutor: func(in ...interface{}) {
                // any thing
            },
            Params: []interface{1, "abc"},
        })
	}
    // wait for job completed
	engine.SendWithGroup(group1...)
    engine.Wait()
```

### safequeue scale up/down

```go
    engine.ScaleUp(5)
    engine.ScaleDown(2)
```


### Graceful shutdown
```go
    jobs := e.quicklyExec.TerminatingHandler()
		if len(jobs) > 0 {
			for _, job := range jobs {
				job.Exectutor(job.Params...)
			}
		}
		log.Print("quick exec job executed: ", len(jobs))
```

## Race

+ For run a task and limit time process

```go
_, err := Race(200*time.Millisecond, func(i ...interface{}) (interface{}, error) {
    time.Sleep(time.Duration(i[0].(int)) * time.Millisecond)
    return nil, nil
}, 30)
log.Print("run: ", time.Since(now))
```

+ With SafeQueue

```go
engine := RunSafeQueue(&SafeQueueConfig{
    NumberWorkers: 3, Capacity: 500,
    WaitGroup: &sync.WaitGroup{},
})
_, err = RaceWithSafeQueue(engine, 200*time.Millisecond, func(i ...interface{}) (interface{}, error) {
    time.Sleep(time.Duration(i[0].(int)) * time.Millisecond)
    return nil, nil
}, 500)
```