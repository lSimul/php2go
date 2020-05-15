package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"

	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/visitor"

	"github.com/lSimul/php2go/p"
)

func main() {
	flag.Parse()
	if len(os.Args) < 2 {
		fmt.Println("Usage: php2go <php file> [<output folder>]")
		return
	}

	p := p.NewParser(p.NewNameTranslator(), p.NewFunctionTranslator())
	gc := p.RunFromString(os.Args[1], true)

	if len(os.Args) < 3 {
		for _, f := range gc.Files {
			fmt.Println(f.String())
		}
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
