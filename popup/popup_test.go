package popup

import "testing"

func newLedger() *Ledger {
	return &Ledger{MergeMs: 500, MaxStack: 3, MaxPopups: 10}
}

func TestFirstEventCreatesPopup(t *testing.T) {
	got := newLedger().Add("e1", "t1", "s1", "hit", 0, 10)
	if got.Merged {
		t.Fatalf("首条不该是合并：%+v", got)
	}
	if got.Amount != 10 || got.Seq != 1 {
		t.Fatalf("首条结果不对：%+v", got)
	}
}

func TestSameKeyWithinWindowMerges(t *testing.T) {
	l := newLedger()
	first := l.Add("e1", "t1", "s1", "hit", 0, 10)
	second := l.Add("e2", "t1", "s1", "hit", 100, 5)
	if !second.Merged {
		t.Fatalf("同键同窗口应合并：%+v", second)
	}
	if second.Seq != first.Seq {
		t.Fatalf("合并后序号应不变：%d vs %d", second.Seq, first.Seq)
	}
	if second.Amount != 15 {
		t.Fatalf("累计金额=%d，期望 15", second.Amount)
	}
}

func TestDifferentTargetGetsNewPopup(t *testing.T) {
	l := newLedger()
	first := l.Add("e1", "t1", "s1", "hit", 0, 10)
	second := l.Add("e2", "t2", "s1", "hit", 0, 10)
	if second.Merged {
		t.Fatalf("不同目标不该合并：%+v", second)
	}
	if second.Seq <= first.Seq {
		t.Fatalf("新条序号应更大：%d vs %d", second.Seq, first.Seq)
	}
}

func TestAmountAccumulates(t *testing.T) {
	l := newLedger()
	l.Add("e1", "t1", "s1", "hit", 0, 3)
	l.Add("e2", "t1", "s1", "hit", 10, 4)
	got := l.Add("e3", "t1", "s1", "hit", 20, 5)
	if got.Amount != 12 {
		t.Fatalf("累计金额=%d，期望 12", got.Amount)
	}
}
