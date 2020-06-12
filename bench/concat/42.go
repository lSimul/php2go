package main

import (
	"flag"
	"fmt"

	"github.com/lSimul/php2go/std"
)

var file *string = flag.String("f", "", "Run designated file.")

type global struct {
	start float64
	c     string
	end   float64
	time  float64
}

func main() {
	g := &global{}
	switch *file {

	case "42.php":
		g.mainFunc0()

	default:
		g.mainFunc0()
	}
}
func (g *global) mainFunc0() {
	g.start = std.Microtime()
	g.c = ""
	for j := 0; j < 30; j++ {
		for k := 0; k < 32000; k++ {
			g.c = std.Concat(g.c, "a")
		}
	}
	g.end = std.Microtime()
	g.time = g.end - g.start
	fmt.Printf("Execution time of script = %f sec.\n", g.time)
}
