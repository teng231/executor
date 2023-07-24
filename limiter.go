package executor

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

type Limiter struct {
	Key   string
	Limit int
	rate  *rate.Limiter
	ctx   context.Context
}

func (l *Limiter) Allow(fn func(i ...interface{}) (interface{}, error), data ...interface{}) (interface{}, error) {
	if l.rate == nil {
		l.rate = rate.NewLimiter(rate.Every(time.Second/time.Duration(l.Limit)), 1)
		l.ctx = context.Background()
	}
	l.rate.Wait(l.ctx)
	return fn(data...)
}
