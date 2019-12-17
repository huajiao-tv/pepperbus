# go-keeper
pepper_bus依赖的go-keeper服务的基本配置

## 功能特点
1. 远程统一管理配置
2. 以订阅的形式，不同的项目获取对应的项目配置（domain）
3. 业务代码中以原子替换的方式更新配置变量，同时调用update函数更新

## 基本操作

服务端配置：　

 - 基础目录　/usr/local/gokeeper/ 配置目录　/usr/local/gokeeper/config  重要

 - 数据库bolt目录　/usr/local/gokeeper/database

 - 日志　/usr/local/gokeeper/log 临时数据　/usr/local/gokeeper/tmp

 - 监听客户端端口 listen = :7000

服务端启动：　gokeeper -f /usr/local/gokeeper/keeper.conf
服务端重启：　curl "http://127.0.0.1:17000/conf/reload?domain={your_domain}"

tips: go-keeper:支持层级结构，但是相同名称的配置文件相同内容，子节点的key会覆盖父节点的key

----------

客户端：

- 通过node下的配置生成struct文件： gokeeper-cli -in /home/zw/working/pepper_bus/src/github.com/johntech-o/pepperbus/etc/dev -out ./tmp

- 客户端链接:　-k 127.0.0.1:7000 -d mydomain -n 127.0.0.1
 - k: 上面keeper布置的服务器和监听的端口
 - d: 想要获取的项目名称(domain)
 - n: 订阅者的ip，即当前的机器的ip

