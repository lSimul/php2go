<?php

function fn(int &$a) {
	$a++;
}

function f(int $a) {
	$a++;
}

$a = 0;

fn($a);
echo $a . "\n";
fn($a);
echo $a . "\n";
fn($a);
echo $a . "\n";
fn($a);
echo $a . "\n";
f($a);
echo $a . "\n";
f($a);
echo $a . "\n";
f($a);
echo $a . "\n";
f($a);
echo $a . "\n";
