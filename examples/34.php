<?php

function f(int $a = 0, int $b = 1, int $c = 2): void {
	echo "$a $b $c\n";
}

f();
f(1);
f(2, 3);
f(2, 3, 4);
