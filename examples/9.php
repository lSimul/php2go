<?php
// Convertable for cycles, easy ones. Go is not that strict
// in cycles, only loop condition has extra restrictions, it
// has to be omitted or the expression has to return bool.

for ($i = 0; $i < 10; $i++) {
	echo $i;
	echo "\n";
}

$i = 0;
for (; $i < 10; $i++) echo $i;

// TODO
//for (;;) break;

// Unreachable code. Go is almost OK with it, it compiles.
// But it is not that cool about this, it shows notice in VS Code.
// TODO
//for (; false; ) ;
// Translated to:
/*
for false {}
*/
