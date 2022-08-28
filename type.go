package executor

import (
	"math/rand"
)

const (
	default_capacity_hub = 2000
	default_numberworker = 3
	E_invalid_priority   = "invalid priority"
	E_exector_required   = "exector required"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!@#$%^&*_=-+?|")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
