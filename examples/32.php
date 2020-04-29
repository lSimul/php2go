<?php

a(3);
echo "\n";

function a(int $a): void {
	if ($a < 0) {
		return;
	}
	echo 'a';
	$a--;
	b($a);
	c($a);
}

function b(int $b): void {
	if ($b < 0) {
		return;
	}
	echo 'b';
	$b--;
	a($b);
	c($b);
}

function c(int $c): void {
	if ($c < 0) {
		return;
	}
	echo 'c';
	$c--;
	a($c);
	b($c);
}
