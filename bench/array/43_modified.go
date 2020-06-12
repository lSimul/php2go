// Updated version of the 43.go
// Instead of using std/arrays,
// go with simple array.

package main

import (
	"flag"
	"fmt"

	"github.com/lSimul/php2go/std"
)

var file *string = flag.String("f", "", "Run designated file.")

type global struct {
	start float64
	TIMES int
	a     []int
	end   float64
	time  float64
}

func main() {
	g := &global{}
	switch *file {

	case "43.php":
		g.mainFunc0()

	default:
		g.mainFunc0()
	}
}
func (g *global) mainFunc0() {
	g.start = std.Microtime()
	g.TIMES = 10000
	g.a = make([]int, 1, g.TIMES)
	g.a[0] = 1
	for i := 1; i < g.TIMES; i++ {
		g.a = append(g.a, i)
	}
	for i := 0; i < len(g.a)-1; i++ {
		for j := i + 1; j < len(g.a); j++ {
			t := g.a[i]
			if g.a[i] < g.a[j] {
				g.a[i] = g.a[j]
				g.a[j] = t
			}
		}
	}
	g.end = std.Microtime()
	g.time = g.end - g.start
	fmt.Printf("Execution time of script = %f sec.\n", g.time)
}
