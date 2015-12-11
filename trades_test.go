package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestTradesToJSON(t *testing.T) {
	trades := NewTrades()
	var entries []Entry = []Entry{{"2001/01/01", "instr", "instrlong", "possession", "violation", "reason", "change", "volume", time.Now().Unix()}}
	trades.NewEntries(entries)

	js, err := json.Marshal(*trades.GetTradesSince(0))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(js))
}

func TestMergeEntries(t *testing.T) {
	e1 := Entry{"2001/01/01", "instr", "instrlong", "possession", "violation", "reason", "change", "volume", time.Now().Unix()}
	e2 := Entry{"2001/01/01", "instr", "instrlong", "possession", "violation", "reason", "change", "volume", time.Now().Unix()}
	e3 := Entry{"2001/01/02", "instr", "instrlong", "possession", "violation", "reason", "change", "volume", time.Now().Unix()}
	e4 := Entry{"2001/01/03", "instr", "instrlong", "possession", "violation", "reason", "change", "volume", time.Now().Unix()}
	if !e1.Equal(e2) {
		t.Fail()
	}
	var entries []Entry = []Entry{e2, e3, e4}

	e5 := Entry{"2001/01/01", "instr2", "instrlong", "possession", "violation", "reason", "change", "volume", time.Now().Unix()}
	e6 := Entry{"2001/01/01", "instr3", "instrlong", "possession", "violation", "reason", "change", "volume", time.Now().Unix()}
	e7 := Entry{"2001/01/01", "instr", "instrlong", "possession", "violation", "reason", "change", "volume", time.Now().Unix()}

	var updated []Entry = []Entry{e5, e6, e7}

	out, num := MergeEntries(entries, updated)
	fmt.Println(len(out), num)
	if len(out) != 5 || num != 2 {
		t.Fail()
	}
}
