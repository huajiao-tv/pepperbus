#!/bin/bash

ip=$(ip addr | grep "10\." | awk '/^[0-9]+: / {}; /inet.*global/ {print gensub(/(.*)\/(.*)/, "\\1", "g", $2)}')

nohup /data/pepperbus/pepperbus -d=pepperbus_test -k=10.143.164.132:7001 -n=$ip:19840 2>/data/pepperbus/log/panic-pepperbus-`date +%Y%m%d%H`.log &
