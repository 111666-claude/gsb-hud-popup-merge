// Package popup 维护战斗飘字台账：把短时间内的伤害累加到同一条飘字上。
package popup

// Entry 是一条活跃飘字。
type Entry struct {
	Seq    int
	Target string
	Source string
	Kind   string
	Amount int
	Stack  int
	AtMs   int64
}

// Result 是一次写入的结果。
type Result struct {
	Seq    int
	Amount int
	Merged bool
}

type key struct {
	target string
	source string
	kind   string
}

// Ledger 是飘字台账。
type Ledger struct {
	MergeMs   int64
	MaxStack  int
	MaxPopups int

	seq     int
	order   []*Entry
	byKey   map[key]*Entry
	seen    map[string]Result
	Dropped int
	Scanned int
}

// Add 写入一条飘字事件，返回该条飘字当前的序号与累计金额。
func (l *Ledger) Add(eventID, target, source, kind string, atMs int64, amount int) Result {
	if l.seen == nil {
		l.seen = map[string]Result{}
	}
	if prev, ok := l.seen[eventID]; ok {
		return prev
	}

	var res Result
	k := key{target: target, source: source, kind: kind}
	if hit := l.byKey[k]; hit != nil && hit.Stack < l.MaxStack && atMs-hit.AtMs <= l.MergeMs {
		hit.Amount += amount
		hit.Stack++
		hit.AtMs = atMs
		res = Result{Seq: hit.Seq, Amount: hit.Amount, Merged: true}
	} else {
		if l.MaxPopups > 0 && len(l.order) >= l.MaxPopups {
			oldest := l.order[0]
			l.order = l.order[1:]
			delete(l.byKey, key{oldest.Target, oldest.Source, oldest.Kind})
			l.Dropped++
		}
		l.seq++
		e := &Entry{Seq: l.seq, Target: target, Source: source, Kind: kind, Amount: amount, Stack: 1, AtMs: atMs}
		l.order = append(l.order, e)
		if l.byKey == nil {
			l.byKey = map[key]*Entry{}
		}
		l.byKey[k] = e
		res = Result{Seq: e.Seq, Amount: e.Amount}
	}

	l.seen[eventID] = res
	return res
}

// Active 返回当前活跃飘字条数。
func (l *Ledger) Active() int { return len(l.order) }
