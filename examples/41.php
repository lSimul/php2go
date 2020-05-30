<?php

$link = mysqli_connect("localhost", "root", "");
mysqli_select_db($link, "test");

$q = "SELECT DISTINCT id FROM `table`";
$r = mysqli_query($link, $q);
/** @var $l array{id: int} */
while ($l = mysqli_fetch_array($r, MYSQLI_ASSOC)) {
	echo $l['id'];
}
