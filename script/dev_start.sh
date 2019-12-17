#!/bin/bash

/usr/local/bin/redis-server /usr/local/etc/redis.conf &

sudo nginx -c  /usr/local/etc/nginx/nginx.conf

/usr/local/opt/php71/sbin/php-fpm -D --fpm-config /usr/local/etc/php/7.1/php-fpm.conf

