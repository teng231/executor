# executor
fast exec task with go and less mem ops


## Why we need executor?

Go with goroutine is the best structure, it's help developer folk a thread easy than other language.
Go have chanel mechanic to connect goroutine(thread) data.
But what problem??

- Fisrt, `go` keywork folk goroutine but go not commit late time. Then you exact your goroutine can run but you don't know when? It's oke fast with small traffic but when server high load you will scare about that, because is slow, and when you create many goroutine is very slow. `waste of time`

- Second, when you create and release goroutine many time, it's with increase memory and CPU time of system, Sometime some developer forgot release, It's can be make app heavy. `waste of resource`

## How to resolve problem??

This's my resolver.

- We create `exact` or `dynamic` number goroutines. Then using `Job` is a unit bring data information to `Worker` to process. Worker no need release and you only create 1 time or reset when you update config.

- Job bring 2 importance field is: `param` and `exector` and you can call them like `exector(param)` and get all result and return on `callback` is optional.
- It's help your app no need create and release goroutine continuity.
- Easy to access and using when coding.

## Disadvance?

- At you know. If you using callback maybe memory leak if you not control bester.
- If buffer chanel full you need wait a little bits for executed job.

## Installtion

Now it's time you can install lib and experience

```bash
go get github.com/teng231/executor
```

## Usage : Interface list
```go
type ISafeQueue interface {
	Info() SafeQueueInfo // engine info
	Close() error // close all anything
	RescaleUp(numWorker uint) // increase worker
	RescaleDown(numWorker uint) error // reduce worker
	Run() // start
	Send(jobs ...*Job) error // push job to hub
	Wait() // keep block thread
	Done() // Immediate stop wait
}
```

### Initial
```go
    engine = CreateSafeQueue(&SafeQueueConfig{
        NumberWorkers: 3,
        Capacity: 500,
        WaitGroup: &sync.WaitGroup{},
    })
    defer engine.Close() // flush engine

    // go engine.Wait() // folk to other thread
    engine.Wait() // block current thread
```
### Send Simple Job
```go
    // simple job
    j := &Job{
        Exectutor: func(in ...interface{}) {
            // any thing
        },
        Params: []interface{1, "abc"}
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

### Send Job complicated
```go
    // wait for job completed
    j := &Job{
        Exectutor: func(in ...interface{}) {
            // any thing
        },
        Params: []interface{1, "abc"},
        Wg: &sync.WaitGroup{},
    }
    engine.Send(j)
    // wait for job run success
    j.Wait()

    // callback handle async
    // you can sync when use with waitgroup
    j := &Job{
        Exectutor: func(in ...interface{}) {
            // any thing
        },
        CallBack: func(out interface{}, err error) {
            // try some thing here
        }
        Params: []interface{1, "abc"}
    }
    engine.Send(j)
```


### Send Job with groups
```go
    // prepaire a group job.
	group1 := make([]*Job, 0)
	for i := 0; i < 10; i++ {
		group1 = append(group1, &Job{
            Exectutor: func(in ...interface{}) {
                // any thing
            },
            Params: []interface{1, "abc"},
            Wg: &sync.WaitGroup{},
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
