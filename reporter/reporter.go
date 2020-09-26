// Package reporter exposes a way to get up to date usage statistics based on
// arbitrary events.
package reporter

import (
	"context"
	"encoding/json"
	"sync"
)

type Event struct {
	Type      string `json:"type,omitempty"`
	AccountID string `json:"account_id"`
}

type Consumer interface {
	Consume(context.Context) <-chan Event
}

func New(ctx context.Context, consumer Consumer) *reporter {
	rep := &reporter{
		m:     sync.RWMutex{},
		stats: make(Stats),
	}

	go func() {
		for m := range consumer.Consume(ctx) {
			rep.m.Lock()

			s, ok := rep.stats[m.AccountID]
			if !ok {
				s = make(AccountStats)
			}

			v := s[m.Type]
			s[m.Type] = v + 1

			rep.stats[m.AccountID] = s

			rep.m.Unlock()
		}
	}()

	return rep
}

type reporter struct {
	m     sync.RWMutex
	stats Stats
}

func (r *reporter) Report() (interface{}, error) {
	r.m.RLock()
	defer r.m.RUnlock()

	bs, err := json.Marshal(r.stats)
	if err != nil {
		return nil, err
	}

	var statsCopy Stats
	if err := json.Unmarshal(bs, &statsCopy); err != nil {
		return nil, err
	}

	return statsCopy, nil
}

type AccountStats map[string]int
type Stats map[string]AccountStats
