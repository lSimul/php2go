<?php

$a = 1;
if ($a) {
	echo "$a is truthy\n";
} else {
	echo "$a is falsy\n";
}

$b = "1 as string";
if ($b) {
	echo "$b is truthy\n";
} else {
	echo "$b is falsy\n";
}

$b = 0.0;
if ($b) {
	echo "$b is truthy\n";
} else {
	echo "$b is falsy\n";
}

$b = "";
if ($b) {
	echo "$b is truthy\n";
} else {
	echo "$b is falsy\n";
}
