package nqueens

import (
	"testing"

	mc "github.com/heyitsanthony/csp/minconflicts"
)

func TestAttackConflicts(t *testing.T) {
	n := 10
	nq := New(n, mustAttack(n))
	tests := []func([]mc.Val){
		// all first row
		func(rows []mc.Val) {
			for i := 0; i < n; i++ {
				rows[i] = mc.Val(0)
			}
		},
		// diagonals
		func(rows []mc.Val) {
			for i := 0; i < n; i++ {
				rows[i] = mc.Val(i)
			}
		},
		func(rows []mc.Val) {
			for i := 0; i < n; i++ {
				rows[n-i-1] = mc.Val(i)
			}
		},
		// all row 1
		func(rows []mc.Val) {
			for i := 0; i < n; i++ {
				rows[i] = mc.Val(1)
			}
		},
		// all row n-1
		func(rows []mc.Val) {
			for i := 0; i < n; i++ {
				rows[i] = mc.Val(n - 1)
			}
		},
		// alternating row
		func(rows []mc.Val) {
			for i := 0; i < n; i++ {
				rows[i] = mc.Val(i % 2)
			}
		},
	}
	for i, tt := range tests {
		tt(nq.rows)
		if len(nq.Conflicts()) == 0 {
			t.Errorf("#%d: expected conflict", i)
		}
	}
}

func TestPointConflicts(t *testing.T) {
	n := 10
	nq := New(n)
	// Force all cols to use row 0.
	for i := 0; i < n; i++ {
		nq.rows[i] = mc.Val(0)
	}

	tests := []struct {
		x    int
		y    int
		incx int
		incy int

		conflicts int
	}{
		{0, 0, 1, 0, 9},
		{0, 0, 0, 1, 0},
		{0, 0, 1, 1, 0},
		{0, 0, -1, 1, 0},
	}
	for i, tt := range tests {
		if c := nq.line(tt.x, tt.y, tt.incx, tt.incy); c != tt.conflicts {
			t.Errorf("#%d: got %d, expected %d", i, c, tt.conflicts)
		}
	}
}

func TestSolve(t *testing.T) {
	n := 10
	tests := [][]Constraint{
		{},
		{mustAttack(n)},
		{mustAttack(n), mustColinear(n)},
		{mustAttack(n), mustAngle(n)},
		{mustColinear(n), mustAngle(n)},
	}
	for i, tt := range tests {
		nq := New(n, tt...)
		for j := 0; !mc.Step(nq); i++ {
			if j > 20000 {
				t.Errorf("#%d: too many steps", i)
				break
			}
		}
		if cs := nq.Conflicts(); len(cs) != 0 {
			t.Errorf("#%d: expected no conflicts, got %v", i, cs)
		}
	}
}

func TestAngleConflicts(t *testing.T) {
	n := 10
	nq := New(n, mustAngle(n))
	if cs := nq.Conflicts(); len(cs) != 0 {
		t.Errorf("expected no conflicts")
	}

	for i := 0; i < n/2; i++ {
		nq.rows[i] = mc.Val(i * 2)
	}
	if cs := nq.Conflicts(); len(cs) == 0 {
		t.Errorf("expected conflicts")
	}

	nq = New(n, mustAngle(n))
	for i := 0; i < n/2; i++ {
		nq.rows[i*2] = mc.Val(i)
	}
	if cs := nq.Conflicts(); len(cs) == 0 {
		t.Errorf("expected conflicts")
	}
}

func mustAttack(n int) Constraint {
	if c, err := NewConstraintAttack(10); err != nil {
		panic(err)
	} else {
		return c
	}
}

func mustColinear(n int) Constraint {
	if c, err := NewConstraintColinear(2, 10); err != nil {
		panic(err)
	} else {
		return c
	}
}

func mustAngle(n int) Constraint {
	if c, err := NewConstraintAngle(2, 10); err != nil {
		panic(err)
	} else {
		return c
	}
}
