<?php

$a = fn();

function fn(): void {
	echo "This should fail, method does not return value.";
}
