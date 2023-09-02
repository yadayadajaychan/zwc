#!/usr/bin/env bash

SCRIPT_PATH="${0%/*}"
if [ "$0" != "$SCRIPT_PATH" ] && [ "$SCRIPT_PATH" != "" ]; then
    cd $SCRIPT_PATH
fi

export GOCOVERDIR="test-coverage"
rm -r test-coverage
mkdir test-coverage

go build -cover -o zwc ../main/main.go

# vanilla encode
for dir in vanilla/*/
do
	source ${dir}/parameters
	./zwc encode -v -m ${dir}/*.mesg -d ${dir}/*.data -c $CHECKSUM -e $ENCODING | diff -q - ${dir}/*.txt
	if [ $? -ne 0 ]
	then
		echo error during vanilla encode
		exit 1
	fi
done

## data from stdin
for dir in vanilla/*/
do
	source ${dir}/parameters
	cat ${dir}/*.data | ./zwc encode -v -m ${dir}/*.mesg -c $CHECKSUM -e $ENCODING | diff -q - ${dir}/*.txt
	if [ $? -ne 0 ]
	then
		echo error during vanilla encode with data from stdin
		exit 1
	fi
done

## message from stdin
for dir in vanilla/*/
do
	source ${dir}/parameters
	cat ${dir}/*.mesg | ./zwc encode -v -d ${dir}/*.data -c $CHECKSUM -e $ENCODING | diff -q - ${dir}/*.txt
	if [ $? -ne 0 ]
	then
		echo error during vanilla encode with message from stdin
		exit 1
	fi
done

# no-message encode
for dir in no-message/*/
do
	source ${dir}/parameters
	./zwc encode -vn -d ${dir}/*.data -c $CHECKSUM -e $ENCODING | diff -q - ${dir}/*.txt
	if [ $? -ne 0 ]
	then
		echo error during no-message encode
		exit 1
	fi
done

## data from stdin
for dir in no-message/*/
do
	source ${dir}/parameters
	cat ${dir}/*.data | ./zwc encode -vn -c $CHECKSUM -e $ENCODING | diff -q - ${dir}/*.txt
	if [ $? -ne 0 ]
	then
		echo error during no-message encode with data from stdin
		exit 1
	fi
done

rm zwc

echo test.sh: all tests passed

go tool covdata percent -i=test-coverage
