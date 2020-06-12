<?php

$start = microtime(true);

$TIMES = 10000;

$a = [0];
for ($i = 1; $i < $TIMES; $i++) {
	array_push($a, $i);
}

for ($i = 0; $i < count($a) - 1; $i++) {
	for ($j = $i + 1; $j < count($a); $j++) {
		$t = $a[$i];
		if ($a[$i] < $a[$j]) {
			$a[$i] = $a[$j];
			$a[$j] = $t;
		}
	}
}

$end = microtime(true);
$time = $end - $start;
printf("Execution time of script = %f sec.\n", $time);
