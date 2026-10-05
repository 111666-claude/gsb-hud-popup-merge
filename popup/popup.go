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

// Ledger 是飘字台账。
type Ledger struct {
	MergeMs   int64
	MaxStack  int
	MaxPopups int

	seq     int
	order   []*Entry
	seen    map[string]int
	Dropped int
	Scanned int
}

// Add 写入一条飘字事件，返回该条飘字当前的序号与累计金额。
func (l *Ledger) Add(eventID, target, source, kind string, atMs int64, amount int) Result {
	if l.seen == nil {
		l.seen = map[string]int{}
	}
	l.seen[eventID] = l.seq

	var hit *Entry
	for _, e := range l.order {
		l.Scanned++
		if e.Target == target {
			hit = e
		}
	}
	if hit != nil {
		hit.Amount += amount
		hit.Stack++
		hit.AtMs = atMs
		return Result{Seq: hit.Seq, Amount: hit.Amount, Merged: true}
	}

	l.seq++
	e := &Entry{Seq: l.seq, Target: target, Source: source, Kind: kind, Amount: amount, Stack: 1, AtMs: atMs}
	l.order = append(l.order, e)
	return Result{Seq: e.Seq, Amount: e.Amount}
}

// Active 返回当前活跃飘字条数。
func (l *Ledger) Active() int { return len(l.order) }
