package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/z7zmey/php-parser/php7"
	"github.com/z7zmey/php-parser/visitor"
)

func main() {
	src := bytes.NewBufferString(`
		<?php
		/** @var int */
		$a; /// PHP parser ignores these comments, even though they have nice '@'.

		$a < 0;
		$a = 0;
		$a = $b === 0 ? 1 : 2;


		/*if ($a = 0) {
			echo $a;
		} else {
			echo $a;
		}*/
	`)

	parser := php7.NewParser(src, "example.php")
	parser.Parse()

	for _, e := range parser.GetErrors() {
		fmt.Println(e)
	}

	visitor := visitor.Dumper{
		Writer: os.Stdout,
		Indent: "",
	}

	rootNode := parser.GetRootNode()
	rootNode.Walk(&visitor)
}
