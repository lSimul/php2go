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
			echo '<table>';
			echo "\n";
			echo '<tr><th>Key</th><th>Value</th>';
			echo "\n";
			foreach ($_GET as $k => $v) {
				echo "<tr><td>$k</td><td>$v</td></tr>\n";
			}
			echo '</table>';
		} else {
			echo '<p>No GET params</p>';
		}
	?>
</body>
</html>
