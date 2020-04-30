<?php

if ($a = 0) {
	echo "True: \$a = 0\n";
} else {
	echo "False: \$a = 0\n";
}

// if (($a = 4) > 0) {
// echo "True: \$a = 0\n";
// } else {
// echo "False: \$a = 0\n";
// }

// if (($a = 3) < ($b = 4)) {
// echo "True: (\$a = 3) < (\$b = 4)\n";
// } else {
// echo "False: (\$a = 3) < (\$b = 4)\n";
// }

// if ($a = $b = $c = 3) {
// echo "True: \$a = \$b = \$c = 3\n";
// } else {
// echo "False: \$a = \$b = \$c = 3\n";
// }

$b = 10;
while ($b--) {
	echo "$b\n";
}

$b = -10;
while ($b++) {
	echo "$b\n";
}
