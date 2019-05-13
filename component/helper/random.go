package helper

import (
	"math/rand"
	"time"
)

func GetRandomInt(n int) int {
	rand.Seed(int64(time.Now().UnixNano()))
	return rand.Intn(n)
}
