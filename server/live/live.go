// SPDX-License-Identifier: AGPL-3.0-or-later

// Package live tells the share pages' open connections when what they show
// changed. Changes go through Redis, so that every server instance hears of
// them and tells its own connections; without Redis they stay in this one.
//
// A change is a key: "user:<uuid>" when a user's sightings or lists changed,
// "competition:<uuid>" when who takes part in a competition did.
package live

import (
	"slices"
	"strings"
	"sync"

	"github.com/klahr/300sf/server/models"
)

const channel = "live"

type watcher struct {
	mu   sync.Mutex
	keys []string
	// refresh gives the keys again after a competition changed, since who
	// takes part may have; nil when they never change.
	refresh func() []string
	ch      chan struct{}
}

var (
	mu       sync.Mutex
	watchers = map[*watcher]bool{}
	listen   sync.Once
)

// Notify says that what the keys stand for changed.
func Notify(keys ...string) {
	for _, key := range keys {
		if models.Cache == nil {
			changed(key)
			continue
		}
		if err := models.Cache.Publish(channel, key).Err(); err != nil {
			changed(key)
		}
	}
}

// Watch returns a channel that is sent to when any of keys changes, and a
// function that stops watching. A burst of changes may arrive as one.
func Watch(keys []string, refresh func() []string) (<-chan struct{}, func()) {
	listen.Do(subscribe)
	w := &watcher{keys: keys, refresh: refresh, ch: make(chan struct{}, 1)}
	mu.Lock()
	watchers[w] = true
	mu.Unlock()
	return w.ch, func() {
		mu.Lock()
		delete(watchers, w)
		mu.Unlock()
	}
}

func subscribe() {
	if models.Cache == nil {
		return
	}
	messages := models.Cache.Subscribe(channel).Channel()
	go func() {
		for m := range messages {
			changed(m.Payload)
		}
	}()
}

func changed(key string) {
	mu.Lock()
	all := make([]*watcher, 0, len(watchers))
	for w := range watchers {
		all = append(all, w)
	}
	mu.Unlock()

	for _, w := range all {
		w.mu.Lock()
		hit := slices.Contains(w.keys, key)
		if hit && w.refresh != nil && strings.HasPrefix(key, "competition:") {
			w.keys = w.refresh()
		}
		w.mu.Unlock()
		if hit {
			select {
			case w.ch <- struct{}{}:
			default:
			}
		}
	}
}
