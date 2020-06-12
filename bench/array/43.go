package main

import (
	"flag"
	"fmt"
	"github.com/lSimul/php2go/std"
	"github.com/lSimul/php2go/std/array"
)

var file *string = flag.String("f", "", "Run designated file.")

type global struct {
	start float64
	TIMES int
	a     array.Int
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
	g.a = array.NewInt(0)
	for i := 1; i < g.TIMES; i++ {
		g.a.Push(i)
	}
	for i := 0; i < g.a.Count()-1; i++ {
		for j := i + 1; j < g.a.Count(); j++ {
			t := g.a.At(array.NewScalar(i))
			if g.a.At(array.NewScalar(i)) < g.a.At(array.NewScalar(j)) {
				g.a.Edit(array.NewScalar(i), g.a.At(array.NewScalar(j)))
				g.a.Edit(array.NewScalar(j), t)
			}
		}
	}
	g.end = std.Microtime()
	g.time = g.end - g.start
	fmt.Printf("Execution time of script = %f sec.\n", g.time)
}
