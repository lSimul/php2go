package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"php2go/p"

	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/php7"
	"github.com/z7zmey/php-parser/visitor"
)

func main() {
	flag.Parse()
	if len(os.Args) < 2 {
		fmt.Println("Usage: php2go <php file> [<output folder>]")
		return
	}
	name := os.Args[1]
	src, err := ioutil.ReadFile(name)
	if err != nil {
		fmt.Println(err)
		return
	}

	parser := php7.NewParser(src, name)
	parser.Parse()

	if errs := parser.GetErrors(); len(errs) > 0 {
		for _, e := range errs {
			fmt.Print(e)
		}
		os.Exit(1)
	}

	rootNode := parser.GetRootNode()

	print(rootNode)

	p := p.NewParser(p.NewNameTranslator(), p.NewFunctionTranslator())
	gc := p.Run(rootNode.(*node.Root), true)

	if len(os.Args) < 3 {
		fmt.Print(gc)
		return
	}

	var writer bytes.Buffer
	writer.WriteString(gc.String())
	b, err := format.Source(writer.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	output := os.Args[2]
	if err := os.Mkdir(output, 0755); err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	if err := ioutil.WriteFile(output+"/main.go", b, 0644); err != nil {
		fmt.Printf("Writing output file: %v\n", err)
		os.Exit(1)
	}
}

func print(n node.Node) {
	visitor := visitor.Dumper{
		Writer: os.Stderr,
		Indent: "",
	}
	n.Walk(&visitor)
}
