#!/bin/bash

for _ in {1..10}
do
	go run 42.go >> go.txt
done

for _ in {1..10}
do
	php -f 42.php >> php.txt
done

for _ in {1..10}
do
	go run 42_modified.go >> modified.txt
done

