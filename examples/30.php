<?php

$a = ['a', 'b', 'c'];

foreach ($a as $k => $v) {
	echo "$k: $v\n";
}

foreach ([0, 1, 2, 3] as $k => $v) {
	echo "$k: $v\n";
}

foreach ([0, 1, 2, 3] as $v) {
	echo ": $v\n";
}
