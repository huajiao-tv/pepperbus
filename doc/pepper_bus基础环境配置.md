# pepper_bus 基础学习

---

## 基础环境搭建（操作系统:ubuntu16.04s）

1 安装 librrd: sudo apt install librrd-dev // todo 修改成
2 安装 pepper_bus: go get github.com/huajiao-tv/pepperbus
2 安装运行 go_keeper: env GIT_TERMINAL_PROMPT=1 go get github.com/huajiao-tv/pepperbus

- 用 pepper_bus 下的 keeper.conf 替换/usr/local 下 gokeeper 的 keeper.conf
  cp /{your_qbus_project}t/src/github.com/huajiao-tv/pepperbus/etc/keeper.conf /usr/local/gokeeper/keeper.conf

* 新建 domain 并复制 dev 下的 conf 文件到新的 domain 下
  mkdir /usr/local/gokeeper/data/config/mydomain
  cp -r /{your_qbus_project}/src/github.com/huajiao-tv/pepperbus/etc/dev/\* /usr/local/gokeeper/data/config/mydomain/
* 启动　/usr/local/gokeeper -f /usr/local/gokeeper/keeper.conf
* 测试 gokeeper 及 mydomain 　
  　　　 curl "127.0.0.1:17000/domain/list"　
  　　　返回　{"error_code":0,"error":"","data":[{"name":"mydomain","version":0}]}

3 安装运行 php_fpm

- 用 pepper_bus 你的 php-fpm.conf 替换原有的配置 cp /{your_qbus_project}/src/github.com/huajiao-tv/pepperbus/etc/php-fpm.conf /etc/php/7.0/fpm/php-fpm.conf
- 重启 php-fpm sudo service php7.0-fpm restart

4 安装运行 redis

- 安装 redis sudo apt install redis-server
- cp /{your_qbus_project}/src/github.com/huajiao-tv/pepperbus/etc/redis.conf /etc/redis/redis.conf
- 重启 redis sudo service redis restart

5 运行 message_go

- /{your_qbus_project}/bin/pepperbus -k 127.0.0.1:7000 -d mydomain -n 127.0.0.1 -ql queue1

---

## pepper_bus 基本流程验证

1 修改消费文件路径

- 配置文件为 /{your_qbus_project}/src/github.com/huajiao-tv/pepperbus/etc/dev/queue.conf
  设置需要测试的队列的消费文件：
  topic_cgi_script_entry map[string]string = queue1-topic1:{your_project_path}/src/github.com/huajiao-tv/pepperbus/example/php/consume.php,queue1-topic2:{your_project_path}/src/github.com/huajiao-tv/pepperbus/example/php/consume.php,queue2-topic3:/home/q/system/hello/index.php 修改为你的消费 php 文件
- 用新文件替换对应的 mydomain 下的配置文件
- 重启 gokeeper

2 运行一次

- **添加文件的 demo:** /服务器路径/src/github.com/huajiao-tv/pepperbus/example/php/product.php
- **确认加入队列成功:** 018/07/18 11:35:43 mux.go:229: [LOG] AddJobs success: queue1/topic1: inTraceId: 1453608781088024576 retrying: false jobs [{"id":1453608781088024577,"inTraceId":145360878108802
- **mux 读取 job 成功:** mux GetJobsFromQueue success: queue1/topic2: outTraceId: 1453608782291789825 retrying: false jobs [{"id":1453608781088024577,"inTraceId":1453608781088024576,"content":"conntent from produce","retrying":0}]
- **mux 分发给 php-fpm 失败:** cgi connect to fail: dial tcp: missing address
  2018/07/18 11:30:03 mux.go:93: [ERROR] mux serve sendToCGI fail: dial tcp: missing address jobs outTraceId: 1453608066795956225
- **加入重试队列**　[LOG] mux save jobs to retry queue success outTraceId: 1453629620091086848
