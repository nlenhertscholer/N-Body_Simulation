#!/bin/bash

for file in `ls | sort -V | grep .txt`
do
	time=$(cut -f 2 $file)
	echo "$time" > $file
done
