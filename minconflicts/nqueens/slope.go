package nqueens

type slope struct {
	x int
	y int
}

// newSlopes computes all possible slopes m that fall on an NxN grid.
func newSlopes(n int) (slopes []slope) {
	for i := 1; i < n; i++ {
		for j := 1; j < n; j++ {
			if gcd(i, j) != 1 {
				// Multiple of some other slope.
				continue
			}
			slopes = append(slopes, slope{i, j})
		}
	}
	return slopes
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
