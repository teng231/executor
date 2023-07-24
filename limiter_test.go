package executor

import (
	"context"
	"log"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func testTime(i int, limiter *rate.Limiter) {
	limiter.Wait(context.Background())
	log.Printf("%f - %d - %v", limiter.Limit(), i, limiter.Allow())
	// time.Sleep(500 * time.Millisecond)
	log.Print("done job ", i)
}
func TestLimiter(t *testing.T) {
	limiter := rate.NewLimiter(rate.Every(time.Second/10), 1)

	for i := 0; i < 100; i++ {
		testTime(i, limiter)
		// if err := limiter.Wait(context.Background()); err != nil {
		// 	log.Print(err)
		// }
		// log.Printf("%f - %d - %v", limiter.Limit(), i, limiter.Allow())
	}

}
