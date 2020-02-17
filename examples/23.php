<?php

$array = ['a', 'b', 'c'];

foreach ($array as $a) {
	echo "$a\n";
}

for ($i = 0; $i < 3; $i++) {
	array_push($array, "$i");
}

echo "\n";

foreach ($array as $k => $v) {
	echo "$k: $v\n";
}
