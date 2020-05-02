<?php

$a = 1;
$b = $c = 2;
$d = 0.0;
$a = 1 + 1;
func(1, 2, 3, 4);

$for = 1;

function ff(): int {
	return -1;
}

function f(): int {
	return +1;
}

function func(int $a, int $b, int $c, int $d) {
	echo $a + $b + $c + $d;
}
