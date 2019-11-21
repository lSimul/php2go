<?php
// This is something I cannot do in Go without issues. I always look back
// and check for a datatype. This is a valid go code. It does not have solid
// PHP alternative.
/*
package main

import "fmt"

func main() {
	a := 0
	{
		a := "1"
		fmt.Print(a) // "1"
	}

	fmt.Print(a) // 0
	a = 2
	fmt.Print(a) // 2
}
*/
// Translator has to be somewhat smart and when the data types does not match,
// define new variable in this block. This might be the first, really undeter-
// mistic behaviour, no alternative which will work every time and be strongly
// typed.

$a = 0;
{
	$a = "1";
	echo $a; // "1"
}
echo $a; // "1"
$a = 2;
echo $a; // 2

// Without strong types Go works as wrong as PHP:
/*
package main

import "fmt"

func main() {
	var a interface{}
	a = 0
	{
		a = "1"
		fmt.Print(a) // "1"
	}
	fmt.Print(a) // "1"
	a = 2
	fmt.Print(a) // 2
}
*/
