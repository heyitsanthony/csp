package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/heyitsanthony/csp/minconflicts"
	"github.com/heyitsanthony/csp/minconflicts/nqueens"
)

func main() {
	attackFl := flag.Bool("constrain-attack", true, "Contrains attacks.")
	colinearFl := flag.Bool("constrain-colinear", false, "Constrains co-linear points.")
	angleFl := flag.Bool("constrain-angle", true, "Constrains points coincident with the origin.")
	retryFl := flag.Int("retry", 0, "Number of iterations before retrying.")
	verboseFl := flag.Bool("verbose", false, "Verbose output.")
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Fprintf(os.Stderr, "usage: %s [flags] <n>\n", os.Args[0])
		os.Exit(1)
	}

	ns, n := flag.Args()[0], 0
	if _, err := fmt.Sscanf(ns, "%d", &n); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse n from %q (%v)", ns, err)
		os.Exit(1)
	}
	maxIter := *retryFl
	if maxIter == 0 {
		// Use some default.
		maxIter = n * n * n / 2
	}

	cs := []nqueens.Constraint{}
	if *attackFl {
		c, err := nqueens.NewConstraintAttack(n)
		if err != nil {
			panic(err)
		}
		cs = append(cs, c)
	}
	if *colinearFl {
		c, err := nqueens.NewConstraintColinear(2, n)
		if err != nil {
			panic(err)
		}
		cs = append(cs, c)
	}
	if *angleFl {
		c, err := nqueens.NewConstraintAngle(2, n)
		if err != nil {
			panic(err)
		}
		cs = append(cs, c)
	}

	nq, i, retries := solve(n, cs, maxIter)
	fmt.Println(nq.String())
	if *verboseFl {
		fmt.Println("iterations:", i, ". retries:", retries)
	}
}

// solve iteratively applies constraints until full CSP is solved.
func solve(n int, cs []nqueens.Constraint, maxIter int) (nq *nqueens.NQueens, i int, retries int) {
	solved := false
	for !solved {
		nq = nqueens.New(n)
		i = 0
		solved = true
		for j := 1; j <= len(cs) && solved; j++ {
			nq.Constrain(cs[:j]...)
			for solved = false; !solved; solved = minconflicts.Step(nq) {
				if i++; i > maxIter {
					retries++
					break
				}
			}
		}
	}
	return nq, i, retries
}
