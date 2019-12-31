####
# 编译镜像
####
FROM golang

# 创建代码编译目录
RUN mkdir -p /go/src/github.com/huajiao-tv/pepperbus

# 设置工作目录
WORKDIR /go/src/github.com/huajiao-tv/pepperbus

# 拷贝代码
COPY ./ /go/src/github.com/huajiao-tv/pepperbus

# 编译
RUN go build

####
# 运行镜像
####
FROM alpine:3.9

# 源优化
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 运行库
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

# 创建程序目录
RUN mkdir -p /data/pepperbus/log

# 工作路径
WORKDIR /data/pepperbus

COPY --from=0 /go/src/github.com/huajiao-tv/pepperbus/pepperbus .

ENV CLUSTER bus_test
ENV KEEPER 127.0.0.1
ENV SERVICE test

EXPOSE 12017
EXPOSE 19840

CMD ["sh", "-c", "/data/pepperbus/pepperbus -n=${HOSTNAME}.${SERVICE}.default.svc.cluster.local:19840 -d=${CLUSTER} -k=${KEEPER}:7000"]
