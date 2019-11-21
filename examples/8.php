<?php
// Similar to #6, but this cannot be resolved by solid data types,
// interface would bring a lot of pain. PHP is fine with this.
{
	$a = "a";
}
$a++;
echo $a; // "b"
