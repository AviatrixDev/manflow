package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

func InitRandGen(config ConfigFile) *rand.Rand {
	if config.Seed == 0 {
		return rand.New(rand.NewSource(time.Now().UnixNano()))
	} else {
		fmt.Println("Seed: " + strconv.FormatInt(int64(config.Seed), 10))
		return rand.New(rand.NewSource(int64(config.Seed)))
	}
}

func GenBytesValue(randGen *rand.Rand) int {
	return randGen.Intn(1000) + 50
}
