#!/bin/bash

ip=$(ip addr | grep "10\." | awk '/^[0-9]+: / {}; /inet.*global/ {print gensub(/(.*)\/(.*)/, "\\1", "g", $2)}')

echo "/data/pepperbus/pepperbus -d=pepperbus_pro -k=inner.gokeeper.huajiao.com:7000 -n=$ip:19840 2>/data/pepperbus/log/panic-pepperbus-`date +%Y%m%d%H`.log"

nohup /data/pepperbus/pepperbus -d=pepperbus_pro -k=inner.gokeeper.huajiao.com:7000 -n=$ip:19840 2>/data/pepperbus/log/panic-pepperbus-`date +%Y%m%d%H`.log &
