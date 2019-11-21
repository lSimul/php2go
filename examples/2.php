<?php

$a = "This example should fail, string to int assignment";

$a = fn();

function fn(): int {
	return -1;
}
