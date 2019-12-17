#!/bin/bash

function localFuncTestClean() {
	exitFlag=$1

	# 停服务
	pepperbus=`ps aux | grep "pepperbus" | grep $TEST_PEPPERBUS_NODE | grep -v grep | awk -F' ' '{print $2}'`
	if [[ $pepperbus != "" ]];then
       		sudo kill $pepperbus
	fi

	php_fpm=`ps aux | grep $TEST_PHP_FPM_CONF_PATH | grep -v grep | awk -F' ' '{print $2}'`
	if [[ $php_fpm != "" ]];then
       		sudo kill $php_fpm
	fi

	redis_server=`ps aux | grep $TEST_REDIS_SERVER_CONF_PATH | grep -v grep | awk -F' ' '{print $2}'`
	if [[ $redis_server != "" ]];then
       		sudo kill $redis_server
	fi

	# 睡一会儿等待清理完毕
	sleep 2
	echo "Local func test clean up !"

	# 根据参数清理完后判断是否退出
	if [[ ! $exitFlag ]]; then
		exit
	fi
}

function runFuncTest() {
	case=$1
	./bin/funcTest -domain=$TEST_PEPPERBUS_DOMAIN \
		-tc=$case \
		-ghost=$TEST_PEPPERBUS_NODE \
		-gadmin=$TEST_PEPPERBUS_ADMIN \
		-jhost=$TEST_JOB_RES_REDIS \
		-rhost=$TEST_REDIS \
		-lpath=$TEST_PEPPERBUS_ERROR_LOG \
		-user=$TEST_LOGIN_USER \
		-pass=$TEST_LOGIN_PASSWORD \
		-jobCount=$TEST_JOB_COUNT \
		-local=true 
}

function localFuncTestInit() {
	CURDIR=`pwd`
	if [ $CURDIR != `pwd $GOPATH"/src/git.huajiao.com/qmessenger/pepperbus"` ];then
		echo "make.sh funcTest local must run in "$GOPATH"/src/git.huajiao.com/qmessenger/pepperbus"
		exit
	fi

	conf=$1

	# 检查配置文件存在性
	if [[ $conf = "" ]];then
		echo "local func test config file not provided"
		exit
	fi

	if [ ! -f $conf ];then
		echo "local func test config file not exist"
		exit
	fi
	source $conf

	# 检查文件存在性
	fileList=($TEST_PHP_FPM_PATH
       	$TEST_PHP_FPM_CONF_PATH 
	$TEST_REDIS_SERVER_PATH
	$TEST_REDIS_SERVER_CONF_PATH
	)
	for i in ${fileList[@]}
	do
		if [ ! -f $i ];then
			echo $i" configured in "$conf" does not exist!"
			exit
		fi
	done
	# 检查目录存在性
	dirList=($TEST_PHP_FPM_WORK_PATH)
	for i in ${dirList[@]}
	do
		if [ ! -d $i ];then
			echo $i" configured in "$conf" does not exist!"
			exit
		fi
	done

	# 保险起见，先清理之前启动过的
	localFuncTestClean "do not exit"

	# 本地启动
	sudo cp $GOPATH"/src/git.huajiao.com/qmessenger/pepperbus/tester/funcTest/consume.php" $TEST_PHP_FPM_WORK_PATH 
	sudo cp -r $GOPATH"/src/git.huajiao.com/qmessenger/pepperbus/sdk" $TEST_PHP_FPM_WORK_PATH 
	sudo chmod 777 $TEST_PHP_FPM_WORK_PATH"consume.php"
	sudo $TEST_PHP_FPM_PATH -y $TEST_PHP_FPM_CONF_PATH -D
	sudo $TEST_REDIS_SERVER_PATH $TEST_REDIS_SERVER_CONF_PATH & 2>&1 > /dev/null
	
	# 检查php-fpm、redis-server正常启动
	php_fpm=`ps aux | grep $TEST_PHP_FPM_CONF_PATH | grep -v grep | wc -l`
	if [ $php_fpm = "0" ];then
		echo "php-fpm start failed!"
		exit
	fi
	redis_server=`ps aux | grep $TEST_REDIS_SERVER_PATH | grep -v grep | wc -l`
	if [ $php_fpm = "0" ];then
		echo "redis-server start failed!"
		localFuncTestClean
		exit
	fi
	
	sleep 3
	# root 运行，因为如果php以domain socket方式启动，用户权限可能无法访问socket文件
	sudo ./bin/pepperbus -n=$TEST_PEPPERBUS_NODE -d=$TEST_PEPPERBUS_DOMAIN -k=$TEST_PEPPERBUS_KEEPER &
	sleep 3
	
	# 检查pepperbus正常启动
	pepperbus=`ps aux | grep "pepperbus" | grep $TEST_PEPPERBUS_NODE | grep -v grep | wc -l`
	if [ $pepperbus = "0" ];then
		echo "pepperbus start failed!"
		localFuncTestClean
		exit
	fi
}

case $1 in
	"test")
		testFiles=`ls *_test.go | grep -v setting_test | grep -v job_test`
		srcFiles=`ls *.go | grep -v test`
		OLD_GOPATH=$GOPATH
		CURDIR=`pwd`
		export GOPATH=$CURDIR"/../../../.."
		echo $GOPATH
		case $2 in
			"")
				go test -v $testFiles $srcFiles
				;;
			"gateway" | "mux" | "mux_server" | "cgi" | "storage")
				go test -v $2"_test.go" $srcFiles

				;;
			*)
				echo "====================================make test usage====================================\n\n"
				echo "./make.sh test will run *_test.go files\n\n"
				echo "./make.sh test component will run component_test.go files\n\n"
				echo "components: gateway mux mux_server cgi storage\n\n"
				echo "This usage is showed when the specified component test code have not been finished!\n\n"
				echo "=======================================================================================\n\n"
				;;
		esac
		GOPATH=$OLD_GOPATH

		exit
		;;
	"ben")
		OLD_GOPATH=$GOPATH
		CURDIR=`pwd`
		OLD_GOBIN=$GOBIN
		export GOPATH=$CURDIR"/../../../.."
		export GOBIN=$CURDIR"/bin/"
		echo $GOPATH
		echo $GOBIN
		go install git.huajiao.com/qmessenger/pepperbus/tester/ben 
		GOPATH=$OLD_GOPATH
		GOBIN=$OLD_GOBIN

		case $2 in
			"lpush" | "rpop" | "lrem" | "ping")
				tc=$2
				ghost=$3
				num=$4
				parallel=$5
				./bin/ben -tc=$tc -ghost=$ghost -num=$num -gonum=$parallel
				;;
			"cgi")
				tc=$2
				cnetwork=$3
				chost=$4
				keepalive=$5
				num=$6
				parallel=$7
				./bin/ben -tc=$tc -chost=$chost -cnetwork=$cnetwork -cgikeep=$keepalive -num=$num -gonum=$parallel
				;;
			*)
				echo "====================================make ben usage====================================\n\n"
				echo "./make.sh ben tc (ghost|chost) num parallel\n\n"
				echo "Only one of ghost and chost should be appear for special ben test."
				echo "tc: lpush rpop lrem ping cgi\n\n"
				echo "ghost: gateway host\n\n"
				echo "chost: cgi host\n\n"
				echo "cnetwork: cgi network(tcp|unix)\n\n"
				echo "num: ben job count per proc\n\n"
				echo "parallel: count of parallel proc\n\n"
				echo "=======================================================================================\n\n"
				;;
		esac

		exit
		;;
	"funcTest")
		OLD_GOPATH=$GOPATH
		CURDIR=`pwd`
		OLD_GOBIN=$GOBIN
		export GOPATH=$CURDIR"/../../../.."
		export GOBIN=$CURDIR"/bin/"
		echo $GOPATH
		echo $GOBIN
		mkdir -p $GOBIN
		go build -o $GOBIN/pepperbus
		go install git.huajiao.com/qmessenger/pepperbus/tester/funcTest 
		cp $GOPATH"/src/git.huajiao.com/qmessenger/pepperbus/tester/funcTest/remote_exec.sh" $GOBIN
		# 清除日志
		sudo rm bin/log/*log* -f

		case $2 in
			"local")
				# 初始化
				localFuncTestInit $3

				if [[ $4 = "normal" || $4 = "nNormal" || $4 = "topicUnsubscribe" || $4 = "jobException" || $4 = "serviceRestart" || $4 = "badJob" || $4 = "topicAddDel" || $4 = "phpEof" ]]; then
					echo "---local test need root authority---"
					runFuncTest $4
				elif [[ $4 = "" ]];then
					echo "---local test need root authority---"
					runFuncTest "normal"
					runFuncTest "nNormal"
					runFuncTest "topicUnsubscribe"
					runFuncTest "jobException"
					runFuncTest "serviceRestart"
					runFuncTest "badJob"
					runFuncTest "topicAddDel"
					runFuncTest "phpEof"
				else
					echo "local test case must be empty or in normal | nNormal | topicUnsubscribe | jobException | badJob | topicAddDel"
				fi

				# 清理
				localFuncTestClean 
				;;
			"normal" | "nNormal" | "topicUnsubscribe" | "jobException" | "badJob" | "topicAddDel"|"serviceRestart")
				tc=$2
				ghost=$3
				gadmin=$4
				jhost=$5
				rhost=$6
				lpath=$7
				user=$8
				pass=$9
				./bin/funcTest -domain=pepperbus_test/test -tc=$tc -ghost=$ghost -gadmin=$gadmin -jhost=$jhost -rhost=$rhost -lpath=$lpath -user=$user -pass=$pass
				;;
			*)
				echo "====================================make funcTest usage====================================\n\n"
				echo "./make.sh funcTest tc ghost jhost rhost lpath user pass | ./make.sh funcTest local tc user pass \n\n"
				echo "local special the funcTest run in local, and all the service will be start by make.sh"
				echo "tc: normal topicUnsubscribe jobException serviceRestart badJob\n\n"
				echo "ghost: gateway host\n\n"
				echo "jhost: redis host for test job output\n\n"
				echo "rhost: redis host\n\n"
				echo "lpath: log path of pepperbus\n\n"
				echo "user: for remote check, user of ghost\n\n"
				echo "pass: for remote check, password of host\n\n"
				echo "=======================================================================================\n\n"
				;;
		esac
		GOPATH=$OLD_GOPATH
		GOBIN=$OLD_GOBIN

		exit
		;;

esac

mkdir -p ./bin/
go build -o ./bin/pepperbus
