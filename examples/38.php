<?php

for ($a = 0; $a < 5; $a++) {
	echo "$a\n";
	for ($b = 0; $b < 5; $b++) {
		echo "\t$a:$b\n";
		if ($b === 3 && $a === 3) {
			break 1;
		} elseif ($b === 10) {
			continue 2;
			break 2;
		}
	}
	echo "$a\n";
}

foreach ([1, 2, 3, 4, 5] as $v) {
	switch ($v) {
	case 2:
		echo "continue 1\n";
		continue 1;
	case 3:
		continue 2;
	}
	echo "$v\n";
}


