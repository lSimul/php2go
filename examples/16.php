<?php

function ff(int &$a) {
	$a++;
}

function f(int $a) {
	$a++;
}

$a = 0;

ff($a);
echo $a . "\n";
ff($a);
echo $a . "\n";
ff($a);
echo $a . "\n";
ff($a);
echo $a . "\n";
f($a);
echo $a . "\n";
f($a);
echo $a . "\n";
f($a);
echo $a . "\n";
f($a);
echo $a . "\n";
