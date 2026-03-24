package events

import (
	"math/rand"
	"sync"
	"time"
)

type session struct {
	kills  int
	deaths int
	mu     sync.RWMutex
}

var (
	// stats: keeps track of ingame players stats
	stats   = make(map[string]*session)
	statsMu sync.Mutex

	// mostvaluable: keeps track of "mvp" player aka player with most kills
	mostvaluable = make(map[string]int)

	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func getOrCreateSession(xuid string) *session {
	statsMu.Lock()
	defer statsMu.Unlock()

	s, ok := stats[xuid]
	if !ok {
		s = &session{}
		stats[xuid] = s
	}
	return s
}
