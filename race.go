package executor

import (
	"context"
	"errors"
	"time"
)

const (
	E_TimeUp = "error_time_up"
)

type RaceFn = func(...interface{}) (interface{}, error)

// RaceFn any action done func return
func Race(duration time.Duration, fn RaceFn, params ...interface{}) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
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

// RaceFn any action done func return
func RaceWithSafeQueue(sq ISafeQueue, duration time.Duration, fn RaceFn, params ...interface{}) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
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
