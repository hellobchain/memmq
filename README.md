# memmq 

memmq 是一个内存消息代理

## 特性

- 内存消息代理
- 支持 HTTP 或 gRPC 传输
- 支持集群
- 支持分片
- 支持代理
- 支持发现
- 自动重试
- 支持 TLS
- 命令行界面
- 交互式提示
- Go 客户端库

如果未指定 TLS 配置，memmq 默认会生成一个自签名证书。

## API

发布消息
```
/pub?topic=string 将负载作为正文发布
```

订阅消息
```
/sub?topic=string 通过 WebSocket 订阅
```

## 架构

- memmq 服务器是具有内存队列的独立服务器，并提供 HTTP API
- memmq 客户端通过发布/订阅到一个或所有服务器来对 memmq 服务器进行分片或集群
- memmq 代理使用 Go 客户端对 memmq 服务器进行集群，并提供统一的 HTTP API

由于这种简单架构，代理和服务器可以被链式组合以构建消息管道。

## 使用方法

### 安装

```shell
go get github.com/hellobchain/memmemmq
```

### 运行服务器

监听 `*:8081`
```shell
memmq.bin
```

设置服务器地址
```shell
memmq.bin --address=localhost:9091
```

启用 TLS
```shell
memmq.bin --cert_file=cert.pem --key_file=key.pem
```

按主题持久化到文件
```shell
memmq.bin --persist
```

使用 gRPC 传输
```shell
memmq.bin --transport=grpc
```

### 运行代理

memmq可以作为代理运行，包括集群、分片和自动重试功能。

集群：向所有 memmq 服务器发布和订阅

```shell
memmq.bin --proxy --servers=127.0.0.1:8081,127.0.0.1:8082,127.0.0.1:8083
```

分片：根据主题将请求发送到单个服务器

```shell
memmq.bin --proxy --servers=127.0.0.1:8081,127.0.0.1:8082,127.0.0.1:8083 --select=shard
```

解析器：使用名称解析器而不是指定服务器 IP

```shell
memmq.bin --proxy --resolver=dns --servers=memmq.proxy.dev
```

### 运行客户端

发布消息

```shell
echo "一条完全任意的消息" | memmq.bin --client --topic=example --publish --servers=localhost:8081
```

订阅消息

```shell
memmq.bin --client --topic=example --subscribe --servers=localhost:8081
``` 

交互模式
```shell
memmq.bin -i --topic=example
```

### 发布消息

通过 HTTP 发布

```
curl -k -d "一条完全任意的消息" "https://localhost:8081/pub?topic=example"
```

### 订阅消息

通过 WebSocket 订阅

```
curl -k -i -N -H "Connection: Upgrade" \
	-H "Upgrade: websocket" \
	-H "Host: localhost:8081" \
	-H "Origin:http://localhost:8081" \
	-H "Sec-Websocket-Version: 13" \
	-H "Sec-Websocket-Key: memmq" \
	"https://localhost:8081/sub?topic=example"
```

## Go 客户端

memmq 提供了一个简单的 Go 客户端

```go
import "github.com/hellobchain/memmemmq/client"
```

### 发布消息

```go
// 向主题 example 发布消息
err := client.Publish("example", []byte(`bar`))
```

### 订阅消息

```go
// 订阅主题 example
ch, err := client.Subscribe("example")
if err != nil {
	return
}

data := <-ch
```

### 创建新客户端

```go
// 默认连接到本地 memmq 服务器 localhost:8081
c := client.New()
```

gRPC 客户端

```go
import "github.com/hellobchain/memmemmq/client/grpc"

c := grpc.New()
```

### 集群

客户端支持集群功能。发布/订阅操作会针对所有服务器执行。

```go
c := client.New(
	client.WithServers("127.0.0.1:8081", "127.0.0.1:8082", "127.0.0.1:8083"),
)
```

### 分片

客户端支持分片功能，类似于 gomemcache。发布/订阅操作会针对单个服务器执行。

```go
import "github.com/hellobchain/memmemmq/client/selector"

c := client.New(
	client.WithServers("127.0.0.1:8081", "127.0.0.1:8082", "127.0.0.1:8083"),
	client.WithSelector(new(selector.Shard)),
)
```

### 解析器

可以使用名称解析器来发现 memmq 服务器的 IP 地址

```go
import "github.com/hellobchain/memmemmq/client/resolver"

c := client.New(
	// 使用 DNS 解析器
	client.WithResolver(new(resolver.DNS)),
	// 将 DNS 名称作为服务器指定
	client.WithServers("memmq.proxy.local"),
)
```