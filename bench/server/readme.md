# Server example

This will be compiled. Passing flags is not easy and turning off the server is also complicated.

So, after doing go build choose correct port and run something like this, for both PHP and Go.

Csv is not that useful, cool part is the "Time taken for tests" present in STDOUT.

`ab -c 1 -n 100000 -e go1.csv 'localhost:8080/4.php?a=1&b=2&c=3' > go1.txt && ab -c 1 -n 100000 -e php1.csv 'localhost:8080/4.php?a=1&b=2&c=3' > php1.txt`

And same for extra ammount of concurent requests:

`ab -c 50 -n 100000 -e go50.csv 'localhost:8080/4.php?a=1&b=2&c=3' > go50.txt && ab -c 50 -n 100000 -e php50.csv 'localhost:8080/4.php?a=1&b=2&c=3' > php50.txt`

I enjoyed it testing with different ports, it is not funny to turn off servers all the time. Or group it by technology and not by ammount of requests.
