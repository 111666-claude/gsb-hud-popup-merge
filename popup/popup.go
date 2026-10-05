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

// mergeKey 是合并键：target、source、kind 三者都相同才算同一条飘字。
type mergeKey struct {
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
	byKey   map[mergeKey][]*Entry
	seen    map[string]Result
	Dropped int
	Scanned int
}

// Add 写入一条飘字事件，返回该条飘字当前的序号与累计金额。
func (l *Ledger) Add(eventID, target, source, kind string, atMs int64, amount int) Result {
	if l.seen == nil {
		l.seen = map[string]Result{}
	}
	if res, ok := l.seen[eventID]; ok {
		return res
	}

	if l.byKey == nil {
		l.byKey = map[mergeKey][]*Entry{}
	}
	k := mergeKey{target: target, source: source, kind: kind}

	var hit *Entry
	for _, e := range l.byKey[k] {
		l.Scanned++
		if e.Stack >= l.MaxStack {
			continue
		}
		d := atMs - e.AtMs
		if d < 0 {
			d = -d
		}
		if d > l.MergeMs {
			continue
		}
		if hit == nil || e.AtMs >= hit.AtMs {
			hit = e
		}
	}

	var res Result
	if hit != nil {
		hit.Amount += amount
		hit.Stack++
		hit.AtMs = atMs
		res = Result{Seq: hit.Seq, Amount: hit.Amount, Merged: true}
	} else {
		l.seq++
		e := &Entry{Seq: l.seq, Target: target, Source: source, Kind: kind, Amount: amount, Stack: 1, AtMs: atMs}
		l.order = append(l.order, e)
		l.byKey[k] = append(l.byKey[k], e)
		res = Result{Seq: e.Seq, Amount: e.Amount}
		for len(l.order) > l.MaxPopups {
			l.evictOldest()
		}
	}

	l.seen[eventID] = res
	return res
}

// evictOldest 淘汰最旧的一条活跃飘字。
func (l *Ledger) evictOldest() {
	old := l.order[0]
	l.order = l.order[1:]
	k := mergeKey{target: old.Target, source: old.Source, kind: old.Kind}
	list := l.byKey[k]
	for i, e := range list {
		if e == old {
			list = append(list[:i], list[i+1:]...)
			break
		}
	}
	if len(list) == 0 {
		delete(l.byKey, k)
	} else {
		l.byKey[k] = list
	}
	l.Dropped++
}

// Active 返回当前活跃飘字条数。
func (l *Ledger) Active() int { return len(l.order) }
