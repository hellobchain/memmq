FROM golang:1.19 as build

ENV GOPROXY=https://goproxy.cn,direct
# 移动到工作目录：/memmq
WORKDIR /memmq
# 将代码复制到容器中
COPY . .
# 编译成二进制可执行文件app
RUN rm -rf go.sum && make
# 移动到用于存放生成的二进制文件的 /build 目录
WORKDIR /build
RUN cp -r /memmq/bin .

FROM alpine:3.14
# 切换软件源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories \
    && apk update \
    && apk add tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone
RUN apk add --update ca-certificates && \
    rm -rf /var/cache/apk/* /tmp/*
WORKDIR /memmq
# 将二进制文件从 /build 目录复制到这里
COPY --from=build /build/bin/memmq.bin /usr/local/bin

# 启动容器时运行的命令
WORKDIR /memmq
CMD memmq.bin





