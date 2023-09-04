#!/usr/bin/env sh

SCRIPT_PATH="${0%/*}"
if [ "$0" != "$SCRIPT_PATH" ] && [ "$SCRIPT_PATH" != "" ]; then
	cd $SCRIPT_PATH
fi

export GOCOVERDIR="test-coverage"
rm -r test-coverage
mkdir test-coverage

# exit if any commands fail
set -e

go build -cover -o zwc ../main/main.go

for dir in vanilla/*/
do
	source ${dir}/parameters

	# encode
	./zwc encode -m ${dir}/*.mesg -d ${dir}/*.data -c $CHECKSUM -e $ENCODING | diff -q - ${dir}/*.txt

	## data from stdin
	cat ${dir}/*.data | ./zwc encode -m ${dir}/*.mesg -c $CHECKSUM -e $ENCODING | diff -q - ${dir}/*.txt

	## message from stdin
	cat ${dir}/*.mesg | ./zwc encode -d ${dir}/*.data -c $CHECKSUM -e $ENCODING | diff -q - ${dir}/*.txt

	# decode
	./zwc decode -t ${dir}/*.txt | diff -q - ${dir}/*.data

	## text from stdin
	cat ${dir}/*.txt | ./zwc decode | diff -q - ${dir}/*.data
done

for dir in no-message/*/
do
	source ${dir}/parameters

	# no-message encode
	./zwc encode -n -d ${dir}/*.data -c $CHECKSUM -e $ENCODING | diff -q - ${dir}/*.txt

	## data from stdin
	cat ${dir}/*.data | ./zwc encode -n -c $CHECKSUM -e $ENCODING | diff -q - ${dir}/*.txt

	# no-message decode
	./zwc decode -t ${dir}/*.txt | diff -q - ${dir}/*.data

	## text froms stdin
	cat ${dir}/*.txt | ./zwc decode | diff -q - ${dir}/*.data
done

rm zwc

echo test.sh: all tests passed

go tool covdata percent -i=test-coverage
