package main

import (
	"bufio"
	"fmt"
	"os"
	"php2go/p"

	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/php7"
	"github.com/z7zmey/php-parser/visitor"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: php2go <php file>")
		return
	}
	name := os.Args[1]
	file, err := os.Open(name)
	if err != nil {
		fmt.Println(err)
		return
	}

	src := bufio.NewReader(file)
	parser := php7.NewParser(src, name)
	parser.Parse()

	for _, e := range parser.GetErrors() {
		fmt.Println(e)
	}

	rootNode := parser.GetRootNode()

	print(rootNode)
	gc := p.Run(rootNode.(*node.Root))
	gc.Print()
}

func print(n node.Node) {
	visitor := visitor.Dumper{
		Writer: os.Stdout,
		Indent: "",
	}
	n.Walk(&visitor)
}
