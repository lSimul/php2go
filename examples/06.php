<?php
// I am not able to translate this right now, oops.
// $a has to be defined above the block. Expected output
// would be something like this:
/*
	var a int
	{
		a = 0
	}

	a++
	fmt.Print(a)
*/
// This looks stupid, really useful it is only in the
// for cycles:
/*
	for ($i = 0; $i < 10; $i++) {}
	echo $i; // 10
 */
{
	$a = 0;
}

$a++;
echo $a;
