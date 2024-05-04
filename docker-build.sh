#!/bin/bash

if [ $# -eq 1 ] ; then
	llvm_url=$1
	docker build --build-arg llvm_url=$llvm_url --tag ddp-spielplatz .
elif [ $# -eq 3 ] ; then
	llvm_url=$1
	cert_path=./$(basename $2)
	key_path=./$(basename $3)
	cp $2 $cert_path
	cp $3 $key_path
	docker build --build-arg llvm_url=$llvm_url --build-arg certpath=$cert_path --build-arg keypath=$key_path --tag ddp-spielplatz .
else
	echo "Usage: docker-build.sh <llvm-archive-path> [cert-path key-path]"
fi
