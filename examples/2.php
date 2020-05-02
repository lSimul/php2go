<?php

$a = "This example should fail, string to int assignment";

$a = ff();

function ff(): int {
	return -1;
}
