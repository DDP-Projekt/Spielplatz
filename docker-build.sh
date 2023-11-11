#!/bin/bash

tar czf ./DDP.tar.gz -C $DDPPATH .

if [ $# -eq 0 ] ; then
	docker build --build-arg ddppath=./DDP.tar.gz  --tag ddp-spielplatz .
elif [ $# -eq 2 ] ; then
	cert_path=./$(basename $1)
	key_path=./$(basename $2)
	cp $1 $cert_path
	cp $2 $key_path
	docker build --build-arg ddppath=./DDP.tar.gz --build-arg certpath=$cert_path --build-arg keypath=$key_path --tag ddp-spielplatz .
else
	echo "Usage: docker-build.sh [cert_path key_path]"
fi
