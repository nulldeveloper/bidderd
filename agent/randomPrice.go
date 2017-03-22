package agent

import (
	"math"
	"math/rand"
	"time"
)

type randomPrice struct {
	percentage float64
	price      float64
}

func round(f float64) float64 {
	return math.Floor(f + .5)
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func (rp *randomPrice) randomPrice() float64 {
	rm := rp.price * 100
	diff := round(rm * rp.percentage)

	low := int(rm - diff)
	high := int(rm + diff)

	randDec := random(low, high)
	decimal := float64(randDec) / 100.0

	return decimal
}
