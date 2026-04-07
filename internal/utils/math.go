package utils

import (
	"math"
	"math/rand"
	"time"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// CalcJoinReward calculates the reward for a player based on their playtime and performance
func CalcJoinReward(seconds, kills, deaths, joinReward int) int {
	var reward int
	switch {
	case seconds <= 18000:
		reward = joinReward
	case seconds <= 72000:
		reward = joinReward * 3
	case seconds <= 180000:
		reward = joinReward * 5
	default:
		reward = joinReward * 8
	}

	kdr := calcKDR(kills, deaths)
	switch {
	case kdr >= 3.0:
		reward += joinReward * 4
	case kdr >= 2.0:
		reward += joinReward * 3
	case kdr >= 1.5:
		reward += joinReward * 2
	case kdr >= 1.0:
		reward += joinReward
	default:
	}

	return reward
}

// calcKDR calculates the kill-death ratio
func calcKDR(kills, deaths int) float64 {
	if deaths == 0 {
		if kills == 0 {
			return 0
		}
		return float64(kills)
	}
	return float64(kills) / float64(deaths)
}

// RandomReward generates a random reward amount
func RandomReward() int {
	v := rng.Float64()
	bias := 9.9
	return int(math.Pow(v, bias)*199500) + 500
}

// safeMulClamp multiplies two int64 values and clamps the result to the range of int64
func safeMulClamp(a, b int64) int64 {
	if a == 0 || b == 0 {
		return 0
	}
	neg := (a < 0) != (b < 0)
	ua := abs(a)
	ub := abs(b)
	if ua > math.MaxInt64/ub {
		if neg {
			return math.MinInt64
		}
		return math.MaxInt64
	}
	res := ua * ub
	if neg {
		return -res
	}
	return res
}

// abs returns the absolute value of an int64
func abs(x int64) int64 {
	if x < 0 {
		if x == math.MinInt64 {
			return math.MaxInt64
		}
		return -x
	}
	return x
}
