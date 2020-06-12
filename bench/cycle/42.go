package main

import (
	"flag"
	"fmt"
	"github.com/lSimul/php2go/std"
)

var file *string = flag.String("f", "", "Run designated file.")

type global struct {
	start float64
	c     int
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
	g.c = 0
	for i := 0; i < 10; i++ {
		for j := 0; j < 32000; j++ {
			for k := 0; k < 32000; k++ {
				g.c++
				if g.c > 50 {
					g.c = 0
				}
			}
		}
	}
	g.end = std.Microtime()
	g.time = g.end - g.start
	fmt.Printf("Execution time of script = %f sec.\n", g.time)
}
