#!/bin/bash

service=pepperbus
target="/data/$service/"

machine_test=(10.208.65.21)
machine_online=()

function upload() {
    ip="$1"
    echo "$ip:upload"
    ssh root@$ip mkdir -p $target
	ls -l $CI_PROJECT_DIR
    rsync -rtv $CI_PROJECT_DIR/$service $CI_PROJECT_DIR/start.sh $CI_PROJECT_DIR/start_test.sh root@$ip:$target
    if (($? != 0)); then
        echo "$ip:rsync failed"
        exit 1
    fi
}

function deploy() {
	arg=$@	
	if [ $arg == 'online' ]; then
		# deploy_online
		exit
	elif [ $arg == 'test' ];then
		deploy_test
	fi
}

function deploy_online() {
	go build
	for ip in "${machine[@]}";
	do
		upload "$ip"
		start_online "$ip"
	done

	echo "all complete"
}

function deploy_test() {
	go build
	for ip in "${machine_test[@]}";
	do
		upload "$ip"
		start_test "$ip"
	done

	echo "all complete"
}

function start_online() {
    ip="$1"
    echo "$ip:start"

	# ssh root@$ip "
	# pkill -9 pepperbus
	# $target/start.sh
	# "

    if (($? != 0)); then
        exit 1
    fi
}

function start_test() {
    ip="$1"
    echo "$ip:start_test"

	ssh root@$ip "
	pkill -9 pepperbus
	$target/start_test.sh
	"

    if (($? != 0)); then
        exit 1
    fi
}

if (($# > 0)); then
    deploy $@
else
    echo "arg is empty"
fi
