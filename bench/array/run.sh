#!/bin/bash

for _ in {1..10}
do
	go run 43.go >> go.txt
done

for _ in {1..10}
do
	php -f 43.php >> php.txt
done

for _ in {1..10}
do
	go run 43_modified.go >> modified.txt
done
