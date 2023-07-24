package executor

import (
	"context"
	"errors"
	"sync"
	"time"
)

const (
	E_TimeUp = "error_time_up"
)

type RaceFn = func(...interface{}) (interface{}, error)

// RaceFn any action done func return
func Race(duration time.Duration, fn RaceFn, params ...interface{}) (interface{}, error) {
	ctx, cancel := context.WithCancel(context.Background())
	var result interface{}
	var err error
	go func() {
		_result, _err := fn(params...)
		result = _result
		err = _err
		cancel()
	}()

	select {
	case <-ctx.Done():
		return result, err
	case <-time.After(duration):
		return nil, errors.New(E_TimeUp)
	}
}

// Race2Fn any action done func return
func Race2Task(fn1, fn2 RaceFn, duration time.Duration, params ...interface{}) (interface{}, error) {
	ctx, cancel := context.WithCancel(context.Background())
	var result interface{}
	var err error
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		wg.Done()
		_result, _err := fn1(params...)
		result = _result
		err = _err
		cancel()
	}()
	go func() {
		wg.Done()
		_result, _err := fn2(params...)
		result = _result
		err = _err
		cancel()
	}()
	wg.Wait()

	select {
	case <-ctx.Done():
		return result, err
	case <-time.After(duration):
		return nil, errors.New(E_TimeUp)
	}
}

// RaceFn any action done func return
func Race2TaskWithSafeQueue(sq ISafeQueue, fn1, fn2 RaceFn, duration time.Duration, params ...interface{}) (interface{}, error) {
	ctx, cancel := context.WithCancel(context.Background())
	var result interface{}
	var err error

	sq.Send(&Job{
		Params: params,
		Exectutor: func(i ...interface{}) (interface{}, error) {
			_result, _err := fn1(params...)
			result = _result
			err = _err
			cancel()
			return _result, _err
		},
	}, &Job{
		Params: params,
		Exectutor: func(i ...interface{}) (interface{}, error) {
			_result, _err := fn2(params...)
			result = _result
			err = _err
			cancel()
			return _result, _err
		},
	})
	select {
	case <-ctx.Done():
		return result, err
	case <-time.After(duration):
		return nil, errors.New(E_TimeUp)
	}
}

// RaceFn any action done func return
func RaceWithSafeQueue(sq ISafeQueue, duration time.Duration, fn RaceFn, params ...interface{}) (interface{}, error) {
	ctx, cancel := context.WithCancel(context.Background())
	var result interface{}
	var err error

	sq.Send(&Job{
		Params: params,
		Exectutor: func(i ...interface{}) (interface{}, error) {
			_result, _err := fn(params...)
			result = _result
			err = _err
			cancel()
			return _result, _err
		},
	})

	select {
	case <-ctx.Done():
		return result, err
	case <-time.After(duration):
		return nil, errors.New(E_TimeUp)
	}
}
