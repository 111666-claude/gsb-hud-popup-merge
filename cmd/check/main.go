// Command check 用固定场景复现飘字合并的判定结果。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"example.com/hud-popup-merge/popup"
)

type req struct {
	ID     string
	Target string
	Source string
	Kind   string
	At     int64
	Amount int
}

type line struct {
	ID     string `json:"id"`
	Seq    int    `json:"seq"`
	Amount int    `json:"amount"`
	Merged bool   `json:"merged"`
}

func emit(l *popup.Ledger, reqs []req) {
	for _, r := range reqs {
		res := l.Add(r.ID, r.Target, r.Source, r.Kind, r.At, r.Amount)
		out, _ := json.Marshal(line{ID: r.ID, Seq: res.Seq, Amount: res.Amount, Merged: res.Merged})
		fmt.Println(string(out))
	}
}

func scaleReqs(count, targets int) []req {
	out := make([]req, 0, count)
	for i := 0; i < count; i++ {
		out = append(out, req{
			ID:     fmt.Sprintf("w%d", i),
			Target: fmt.Sprintf("t%d", i%targets),
			Source: "s1",
			Kind:   "hit",
			At:     int64(i),
			Amount: 1,
		})
	}
	return out
}

func main() {
	sample := flag.String("sample", "cross-source", "场景名")
	flag.Parse()

	var l *popup.Ledger
	var reqs []req

	switch *sample {
	case "cross-source":
		l = &popup.Ledger{MergeMs: 1000, MaxStack: 9, MaxPopups: 9}
		reqs = []req{
			{ID: "e1", Target: "t1", Source: "s1", Kind: "hit", At: 0, Amount: 10},
			{ID: "e2", Target: "t1", Source: "s2", Kind: "hit", At: 10, Amount: 7},
		}
	case "window":
		l = &popup.Ledger{MergeMs: 300, MaxStack: 9, MaxPopups: 9}
		reqs = []req{
			{ID: "e1", Target: "t1", Source: "s1", Kind: "hit", At: 0, Amount: 10},
			{ID: "e2", Target: "t1", Source: "s1", Kind: "hit", At: 400, Amount: 7},
		}
	case "stack":
		l = &popup.Ledger{MergeMs: 100000, MaxStack: 2, MaxPopups: 9}
		reqs = []req{
			{ID: "e1", Target: "t1", Source: "s1", Kind: "hit", At: 0, Amount: 10},
			{ID: "e2", Target: "t1", Source: "s1", Kind: "hit", At: 10, Amount: 5},
			{ID: "e3", Target: "t1", Source: "s1", Kind: "hit", At: 20, Amount: 3},
		}
	case "evict":
		l = &popup.Ledger{MergeMs: 1, MaxStack: 9, MaxPopups: 2}
		reqs = []req{
			{ID: "e1", Target: "t1", Source: "s1", Kind: "hit", At: 0, Amount: 1},
			{ID: "e2", Target: "t2", Source: "s1", Kind: "hit", At: 10, Amount: 2},
			{ID: "e3", Target: "t3", Source: "s1", Kind: "hit", At: 20, Amount: 3},
		}
	case "dup":
		l = &popup.Ledger{MergeMs: 1000, MaxStack: 9, MaxPopups: 9}
		reqs = []req{
			{ID: "e1", Target: "t1", Source: "s1", Kind: "hit", At: 0, Amount: 10},
			{ID: "e1", Target: "t1", Source: "s1", Kind: "hit", At: 10, Amount: 10},
		}
	case "work":
		l = &popup.Ledger{MergeMs: 1000000000, MaxStack: 100000, MaxPopups: 100000}
		reqs = scaleReqs(3000, 300)
	default:
		fmt.Fprintf(os.Stderr, "未知场景 %q\n", *sample)
		os.Exit(2)
	}

	if *sample != "work" {
		emit(l, reqs)
	} else {
		for _, r := range reqs {
			l.Add(r.ID, r.Target, r.Source, r.Kind, r.At, r.Amount)
		}
	}

	switch *sample {
	case "evict":
		out, _ := json.Marshal(map[string]int{"dropped": l.Dropped, "active": l.Active()})
		fmt.Println(string(out))
	case "work":
		out, _ := json.Marshal(map[string]int{"scanned": l.Scanned, "popups": l.Active()})
		fmt.Println(string(out))
	}
}
