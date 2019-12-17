# Pepper Bus Service

### 1. Getting started
To get started,and run a demo. you’ll first need to install the following dependencies, which are used to compile the backend and frontend respectively.

- **Go >= 1.9**
Then, run the following to download frontend and backend dependencies:
- **php >= 7.1**

### 2. Compiling

```
go get github.com/johntech-o/pepperbus/
dep ensure //  brew install dep  
go install github.com/johntech-o/pepperbus/
```


### 3. Development
- start gokeeper service

```
mkdir -p /usr/local/gokeeper

cp your_path/github.com/johntech-o/pepperbus/etc/keeper.conf /usr/local/gokeeper/keeper.conf

```
- gokeeper load dev config of pepperbus to serve 

```
rm -rf /usr/local/gokeeper/data/database/keeper.db 
cp your_path/github.com/johntech-o/pepperbus/etc/dev/*  /usr/local/gokeeper/data/config/dev/

gokeeper -f /usr/local/gokeeper/keeper.conf
```
- install redis
- start redis service

```
mv your_path/github.com/johntech-o/pepperbus/etc/redis.conf /usr/local/etc/redis.conf

/usr/local/bin/redis-server /usr/local/etc/redis.conf
```
- install php
- start php-fpm service

```
php-fpm --nodaemonize --fpm-config your_path/github.com/johntech-o/pepperbus/etc/php-fpm.conf
```
- start pepperbus 

```
pepperbus -d dev -k=127.0.0.1:7000 -n=127.0.0.1:test -ql=queue1
```

- restart pepperbus 

```
//default admin port is 19840 
curl "127.0.0.1:19840/manage/service?op=restart" --cookie "manage=true"
```

- stop pepperbus 

```
//default admin port is 19840 
curl "127.0.0.1:19840/manage/service?op=stop" --cookie "manage=true"
```

### 4. Test U PepperBus in develop environment
- create a php test env

```
mkdir -p /data/exapmle/
ln -s  your_path/github.com/johntech-o/example/php  /data/example/php
```
- check the config in gokeeper

```
// check the gokeeper config is ok for test 
curl http://127.0.0.1:17000/conf/list?domain=dev
```

### 5. Test

#### 5.1 Unit Test

```
./make.sh test $COMPONENT will run unit test of $COMPONENT
$COMPONENT in values of gateway|mux_server|mux|storage|cgi
./make test will run all unit test of components above
```
#### 5.2 Function Test

```
./make.sh funcTest local $CONFIG_FILE $TEST_CASE
Local function Test will start all the services of pepperbus due to infos in $CONFIG_FILE, and run the function test of $TEST_CASE
$TEST_CASE must be empty or in values of normal|topicUnsubscribe|jobException|badJob|topicAddDel


./make.sh funcTest $TEST_CASE $GHOST $JHOST $RHOST $LPATH $USER $PASSWORD
$TEST_CASE in values of normal|topicUnsubscribe|jobException|badJob|topicAddDel
$GHOST is the gateway host with port of pepperbus
$JHOST is the job exec host of pepperbus
$RHOST is the redis host with port and authorization of pepperbus
$LPATH is the error log path of pepperbus
$USER is the remote user of ghost for remote log check
$PASSWORD is the remote user password of ghost for remote log check
```
#### 5.2 Benchmark Test

```
./make.sh ben lpush $GHOST $NUM $PARALLEL
$GHOST is the gateway host with port of pepperbus
$NUM job count in one proc
$PARALLEL parallel procs count
```
