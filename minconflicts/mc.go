package minconflicts

import "math/rand"

// Var/Val can be any type in a CSP, but assume there's some
// fast integer indexing/encoding to avoid the overhead of
// needing interface{}s here.

type Var int
type Val int

type Conflict struct {
	Var       Var
	Conflicts int
}

// MinConflict implementers can solve constraint satisfaction problems
// by stepping through the min-conflicts repair method.
type MinConflict interface {
	Conflicts() []Conflict
	// Heuristic returns the best value choice for a given variable
	// along with a score of the cost of the choice.
	Heuristic(Var) (Val, int)
	// Assign maps a value to a variable.
	Assign(Var, Val)
	// Unassign removes the assignment from a variable.
	Unassign(Var) Val
	// Size returns the number of variables.
	Size() int
}

// Step runs a single iteration of the heuristic repair method. It
// returns true iff the problem is solved.
func Step(mc MinConflict) bool {
	cs := mc.Conflicts()
	if len(cs) == 0 {
		return true
	}
	// Randomly select conflicting variable.
	off := rand.Intn(len(cs))
	for i := range cs {
		c := cs[(i+off)%len(cs)]
		// Assign variable to value with fewest conflicts.
		val := mc.Unassign(c.Var)
		minVal, minCost := mc.Heuristic(c.Var)
		if minCost < c.Conflicts {
			// Improved variable's conflict count.
			mc.Assign(c.Var, minVal)
			return false
		}
		mc.Assign(c.Var, val)
		// Could not improve variable, try next.
	}
	// No change; swap constraints to break local minima.
	var1 := cs[off].Var
	if rand.Intn(2) == 0 {
		var1 = Var(rand.Intn(mc.Size()))
	}
	var2 := var1
	for var2 == var1 {
		var2 = Var(rand.Intn(mc.Size()))
	}
	val1, val2 := mc.Unassign(var1), mc.Unassign(var2)
	mc.Assign(var1, val2)
	mc.Assign(var2, val1)
	return false
}
