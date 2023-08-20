#!/usr/bin/env bash

# vanilla encode
for dir in */
do
	source ${dir}/parameters
	go run ../main/main.go encode -v -m ${dir}/*.mesg -d ${dir}/*.data -c $CHECKSUM -e $ENCODING | diff -q - ${dir}/*.txt
	if [ $? -ne 0 ]
	then
		echo error during vanilla encode
		exit 1
	fi
done
