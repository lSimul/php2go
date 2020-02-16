<?php

$array = [0];
for ($i = 1; $i <= 5; $i++) {
	array_push($array, $i);
}

for ($i = 0; $i < 6; $i++) {
	echo $array[$i];
	echo "\n";
}
