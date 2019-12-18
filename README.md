# pepperbus


pepperbus 
是一个实时的、分布式的消息队列服务，具有持久化的任务与消息队列、任务的调度处理框架、任务的跟踪与管理。

## Features

- 支持redis协议作为总线数据输入，充分利用业务端成熟的redis扩展，对业务端友好，迁移和改造成本低
- 通过fastcgi协议与后端通信，与实际的业务逻辑解耦，方便的支持php和golang等多种语言接入
- 数据的输入和输出对于长连接友好，解决LNMP解决方案，长连接情况下，由于集群连接过多给后端服务造成压力，通过代理方式与最终的后端服务做连接复用，对输入不限制长连，对输出维持有限长连接
- 总线队列支持常规的生产者和消费者模式，消息可以按订阅关系重复消费
- 全链路监控策略，可以获取流水线作业的操作序列，完整自描述其关系
- 队列拥塞状态监控，通过与消费者协议交互，反馈处理异常的队列，计算队列消费者处理能力，动态扩容消费进程
- 总线支持定时任务的设置和管理，动态控制和迁移定时任务，解决常规的crontab的管理问题
- 总线队列支持平滑重启，平稳恢复所有任务

## 架构描述

- 业务服务器可以与bus总线实例混合部署也可以独立部署，混合部署时可以优先访问本地总线服务器，独立分离部署可以与总线服务用域名进行通信
- 总线服务器与php-fpm交互可以通过本地也可以通过网络
- 总体架构图：
    
![pepper_bus_overview](http://static.s3.huajiao.com/Object.access/hj-video/cGVwcGVyX2J1c19vdmVydmlldy5wbmc=)

## 组件

- dashboard: 提供用户管理的web ui。
- gokeeper: 提供配置管理服务。

## 项目依赖

- Docker 18.09+ ([installation manual](https://docs.docker.com/install))
- Docker-compose 1.23.2+ ([installation manual](https://docs.docker.com/compose/install/)) 


## 快速启动

- 克隆代码仓库

```bash
$ git clone https://github.com/huajiao-tv/pepperbus.git
$ git clone https://github.com/qmessenger/job-dashboard.git 
#用户管理页面
```


- 启动pepperbus与dashboard服务

进入pepperbus根目录，执行

```bash
$ cd deploy
$ docker-compose up -d
```

通过上述命令，pepperbus相关服务均启动完毕，您可以从通过 
http://127.0.0.1:9528 
访问pepperbus用户管理页面, 默认管理员账号 admin:admin。

> 注意：服务启动后还需要配置集群和存储等信息才可正常使用。详见 
[**用户使用说明**](doc/user_instruction.md)


## 文档

- [总体设计](doc/design.md)
- [用户使用说明](doc/user_instruction.md)
- [SDK使用说明](doc/sdk/sdk_instruction.md)


