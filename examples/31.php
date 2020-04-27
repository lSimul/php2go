<?php

$array = ['placeholder'];
$array['e'] = '5';

if (isset($array['e'])) {
	echo "\"e\": $array[e]\n";
} else {
	echo "\"e\" is not set.\n";
}

unset($array['e']);

if (isset($array['e'])) {
	echo "\"e\": $array[e]\n";
} else {
	echo "\"e\" is not set.\n";
}

$array['e'] = 'f';
if (isset($array['e'])) {
	echo "\"e\": $array[e]\n";
} else {
	echo "\"e\" is not set.\n";
}
