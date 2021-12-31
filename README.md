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

## Installtion

Now it's time you can install lib and experience

```bash
go get github.com/teng231/executor
```
