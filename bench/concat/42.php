<?php

$start = microtime(true);

$c = '';
for ($j = 0; $j < 30; $j++) {
	for ($k = 0; $k < 32000; $k++) {
		$c .= 'a';
	}
}

$end = microtime(true);
$time = $end - $start;
printf("Execution time of script = %f sec.\n", $time);
