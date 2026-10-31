// SPDX-License-Identifier: AGPL-3.0-or-later
package live

import (
	"testing"
	"time"
)

func received(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	case <-time.After(100 * time.Millisecond):
		return false
	}
}

// Without Redis, changes reach the watchers in this process.
func TestWatchersHearTheirKeys(t *testing.T) {
	anna, stopAnna := Watch([]string{"user:anna"}, nil)
	defer stopAnna()
	race, stopRace := Watch([]string{"competition:race", "user:erik"}, func() []string {
		return []string{"competition:race", "user:erik", "user:maja"}
	})
	defer stopRace()

	Notify("user:erik")
	if received(anna) || !received(race) {
		t.Error("user:erik reached the wrong watcher")
	}
	// Maja isn't watched until the competition says someone joined.
	Notify("user:maja")
	if received(race) {
		t.Error("user:maja heard before she joined")
	}
	Notify("competition:race")
	if !received(race) {
		t.Error("competition change not heard")
	}
	Notify("user:maja")
	if !received(race) {
		t.Error("user:maja not heard after joining")
	}

	stopAnna()
	Notify("user:anna")
	if received(anna) {
		t.Error("heard after stopping")
	}
}

// A refresh that fails keeps the keys it had, so the page still hears them.
func TestFailedRefreshKeepsKeys(t *testing.T) {
	race, stop := Watch([]string{"competition:failing", "user:olle"}, func() []string { return nil })
	defer stop()

	Notify("competition:failing")
	if !received(race) {
		t.Error("competition change not heard")
	}
	Notify("user:olle")
	if !received(race) {
		t.Error("user:olle not heard after a failed refresh")
	}
}
