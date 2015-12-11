package main

import (
	//	"github.com/ant0ine/go-json-rest/rest"
	"fmt"
	"log"
	//"strconv"
	"sync"
	"time"
)

/* Remember that names must begin with Upper case letter to be exported into json */

type Entry struct {
	Date       string `json:date`
	Instr      string `json:instr`
	InstrLong  string `json:instrlong`
	Possession string `json:possession`
	Violation  string `json:violation`
	Reason     string `json:reason`
	Change     string `json:change`
	Volume     string `json:volume`
	FirstSeen  int64  `json:seen`
}

func (e Entry) String() string {
	return fmt.Sprintf("%s %s %s %s %s %s %s %s", e.Date, e.Instr, e.InstrLong, e.Possession, e.Violation, e.Reason, e.Change, e.Volume)
}

func (e Entry) Equal(f Entry) bool {
	return (e.Date == f.Date) && (e.Instr == f.Instr) && (e.Possession == f.Possession) && (e.Change == f.Change) && (e.Volume == f.Volume)
}

type TradeResponse struct {
	Entries []Entry `json:entries`
	Now     int64   `json:now`
}

func MakeTradeResponse(entries []Entry) *TradeResponse {
	return &TradeResponse{entries, time.Now().Unix()}
}

type Trades struct {
	lock     sync.RWMutex
	entries  []Entry
	children []chan struct{}
}

func NewTrades() *Trades {
	return &Trades{entries: make([]Entry, 0)}
}

// TODO: feed this will callbacks instead..
func (t *Trades) StartBackground(refreshRate time.Duration) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.children = append(t.children, aktietorget_background(t, refreshRate))
}

func (t *Trades) GetTradesSince(since int64) *TradeResponse {
	t.lock.RLock()

	out := make([]Entry, 0, len(t.entries))
	for _, e := range t.entries {
		if e.FirstSeen > since {
			out = append(out, e)
		}
	}

	t.lock.RUnlock()

	return MakeTradeResponse(out)
}

func (t *Trades) Stop() {
	for _, c := range t.children {
		close(c)
	}
}

func MergeEntries(orig []Entry, new []Entry) ([]Entry, int) {

	addedEntries := make([]Entry, 0, len(new))

	for ni := 0; ni < len(new); ni++ {
		added := true
		for oi := 0; oi < len(orig) && added; oi++ {
			if orig[oi].Equal(new[ni]) {
				added = false
			}
		}
		if added {
			addedEntries = append(addedEntries, new[ni])
		}
	}
	return append(addedEntries, orig...), len(addedEntries)
}

func (t *Trades) NewEntries(entries []Entry) {
	t.lock.Lock()
	defer t.lock.Unlock()

	merged, num := MergeEntries(t.entries, entries)
	log.Println("Added ", num)

	t.entries = merged
}
