# PepperBus Docker 简要说明

只用于 Docker For Mac

## QuickStart

1.  升级 Docker 到最新版

2.  修改 docker-compose.yaml 将 php-fpm 和 gokeeper 中 volumes 第一个值 `/Users/specode/Code/pepperbus/` 修改为你的 PepperBus 路径

3.  首次执行 `docker-compose up -d`  

4.  修改或更新项目代码后，执行 `docker-compose up --build`

## 补充说明

1.  `docker exec -it pepperbus_main /bin/bash` 可进入 PepperBus 主程序所在的实例，可查看日志

2.  `docker exec -it pepperbus_php-fpm /bin/bash` 可进入 PHP-FPM 所在实例，可执行简版 SDK

3.  `docker exec -it pepperbus_redis redis-cli` 可进入 Redis
