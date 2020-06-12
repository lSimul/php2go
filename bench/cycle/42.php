<?php

$start = microtime(true);

$c = 0;
for ($i = 0; $i < 10; $i++) {
	for ($j = 0; $j < 32000; $j++) {
		for ($k = 0; $k < 32000; $k++) {
			$c++;
			if ($c > 50) {
				$c = 0;
			}
		}
	}
}

$end = microtime(true);
$time = $end - $start;
printf("Execution time of script = %f sec.\n", $time);
