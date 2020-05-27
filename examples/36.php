<?php

$a = 0;
echo "$a\n";
inc();
echo "$a\n";

$b = 0;
echo "$b\n";
inc();
echo "$b\n";


function inc(): void {
	global $a;
	$a++;
	// PHP Notice: Undefined variable: b
	// $b++;

	$b = 1;
}
