<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta http-equiv="X-UA-Compatible" content="ie=edge">
	<title>3</title>
</head>
<body>
	<?php
		if (count($_GET) > 0) {
			echo '<ul>';
			foreach ($_GET as $v) {
				echo "<li>$v</li>\n";
			}
			echo '</ul>';
		} else {
			echo '<p>No GET params</p>';
		}
	?>
</body>
</html>
