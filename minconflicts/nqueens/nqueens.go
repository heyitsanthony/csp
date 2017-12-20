package nqueens

import (
	"errors"
	"math/rand"
	"sync"

	mc "github.com/heyitsanthony/csp/minconflicts"
)

var ErrNoSolution = errors.New("no solution")

// NQueens holds state related to an n-queens problem.
type NQueens struct {
	// rows is indexed by column/var to get the assigned row.
	rows []mc.Val
	// n is the size of the board
	n int
	// constraints is the set of constraints to satisfy.
	constraints []Constraint
}

// New creates a new NQueens instance.
func New(n int, cs ...Constraint) *NQueens {
	return &NQueens{
		// Begin with inconsistent assignment; all cols assigned to row 0.
		rows:        make([]mc.Val, n),
		n:           n,
		constraints: cs,
	}
}

// Constrain replaces the constraint set.
func (nq *NQueens) Constrain(cs ...Constraint) { nq.constraints = cs }

func (nq *NQueens) Conflicts() []mc.Conflict {
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(nq.n)
	cs := make([]mc.Conflict, 0, 4)
	for i := 0; i < nq.n; i++ {
		go func(col int) {
			if c := nq.conflicts(col, int(nq.rows[col])); c != 0 {
				mu.Lock()
				cs = append(cs, mc.Conflict{Var: mc.Var(col), Conflicts: c})
				mu.Unlock()
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	return cs
}

func (nq *NQueens) Heuristic(col mc.Var) (mc.Val, int) {
	var wg sync.WaitGroup
	conflicts := make([]int, nq.n)
	// Compute conflicts for all possible value choices.
	for i := 0; i < nq.n; i++ {
		wg.Add(1)
		go func(row int) {
			conflicts[row] = nq.conflicts(int(col), row)
			wg.Done()
		}(i)
	}
	wg.Wait()
	// Find values with minimum conflicts.
	minVal, minConflicts, c := mc.Val(0), conflicts[0]+1, 0
	for row := range conflicts {
		if minConflicts >= conflicts[row] {
			if minConflicts == conflicts[row] {
				c++
			} else {
				c = 1
				minVal, minConflicts = mc.Val(row), conflicts[row]
			}
		}
	}
	if c == 1 {
		// One value with minimum conflict count.
		return minVal, minConflicts
	}
	// Break tie randomly.
	idx := rand.Intn(c)
	for row := range conflicts {
		if conflicts[row] == minConflicts {
			if idx == 0 {
				minVal = mc.Val(row)
				break
			}
			idx--
		}
	}
	return minVal, minConflicts
}

func (nq *NQueens) Assign(v mc.Var, vv mc.Val) { nq.rows[v] = vv }

func (nq *NQueens) Unassign(v mc.Var) mc.Val {
	ret := nq.rows[v]
	nq.rows[v] = -1
	return ret
}

func (nq *NQueens) Size() int { return nq.n }

func (nq *NQueens) String() (ret string) {
	for y := 0; y < nq.n; y++ {
		ret += "|"
		for x := 0; x < nq.n; x++ {
			if nq.rows[x] == mc.Val(y) {
				ret += "Q"
			} else {
				ret += "_"
			}
		}
		ret += "|\n"
	}
	return ret
}

// conflicts counts the number of pieces on the board that will
// that conflict with a piece placed at (x,y).
func (nq *NQueens) conflicts(x, y int) (ret int) {
	for _, constraint := range nq.constraints {
		ret += constraint(nq, x, y)
	}
	return ret
}

// line counts all pieces coincident with (x+t*incx,y+t*incy) where t!=0.
func (nq *NQueens) line(x, y, incx, incy int) (ret int) {
	return nq.ray(x, y, incx, incy) + nq.ray(x, y, -incx, -incy)
}

// ray counts all pieces coincident with (x+t*incx,y+t*incy) where t>1.
func (nq *NQueens) ray(x, y, incx, incy int) (ret int) {
	for i, j := x+incx, y+incy; i >= 0 && i < nq.n && j >= 0 && j < nq.n; i, j = i+incx, j+incy {
		if nq.rows[i] == mc.Val(j) {
			ret++
		}
	}
	return ret
}

type Constraint func(nq *NQueens, x, y int) int

// NewConstraintAttack creates constraints on queen attack moves.
func NewConstraintAttack(n int) (Constraint, error) {
	if n == 2 || n == 3 {
		return nil, ErrNoSolution
	}
	return func(nq *NQueens, x, y int) int {
		// Compute conflicts for horizontal, vertical, diagonal.
		return nq.line(x, y, 1, 0) + nq.line(x, y, 0, 1) +
			nq.line(x, y, 1, 1) + nq.line(x, y, -1, 1)
	}, nil
}

// NewConstraintColinear creates a constraint that forces
// at most maxPts to be on any line for an NxN grid.
func NewConstraintColinear(maxPts, n int) (Constraint, error) {
	// Catch low-value special cases.
	switch n {
	case 1:
		return func(nq *NQueens, x, y int) int { return 0 }, nil
	case 2, 3, 5, 6, 7:
		if maxPts <= 2 {
			return nil, ErrNoSolution
		}
	}
	// Limit to 1+n/maxPts to avoid lines that can't have maxPts points on board.
	// Also ignore the diagonal slope (1,1).
	slopes := newSlopes(1 + n/maxPts)[1:]
	// Compute slopes for maxPt-max lines; given coord is implicitly counted.
	// If maximum points is 2, expect at most 1 other point.
	maxPts--
	return func(nq *NQueens, x, y int) (ret int) {
		for _, s := range slopes {
			// Positive and negative slopes.
			if pts := nq.line(x, y, s.x, s.y); pts > maxPts {
				ret += pts - maxPts
			}
			if pts := nq.line(x, y, -s.x, s.y); pts > maxPts {
				ret += pts - maxPts
			}
		}
		return ret
	}, nil
}

// NewConstraintAngle creates a constraint that forces
// at most maxPts to be on any line starting from the origin.
func NewConstraintAngle(maxPts, n int) (Constraint, error) {
	if n == 1 {
		return func(nq *NQueens, x, y int) int { return 0 }, nil
	}
	slopes := newSlopes(1 + n/maxPts)[1:]
	maxPts--
	return func(nq *NQueens, x, y int) int {
		if x == 0 && y == 0 {
			// Origin must compute all slopes.
			ret := 0
			for _, s := range slopes {
				if pts := nq.line(0, 0, s.x, s.y); pts > maxPts {
					ret += pts - maxPts
				}
			}
			return ret
		}
		if x == 0 || y == 0 {
			// Non-origin axis counted by ordinary n-queens.
			return 0
		}

		d := gcd(x, y)
		pts := nq.line(0, 0, x/d, y/d)
		if nq.rows[0] == 0 {
			// (0,0) is excluded since it's the starting point.
			pts++
		}
		if nq.rows[x] == mc.Val(y) {
			// Want to exclude (x,y) instead of (0,0).
			pts--
		}
		if pts > maxPts {
			return pts - maxPts
		}
		return 0
	}, nil
}
