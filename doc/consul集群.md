Docker部署consul集群

部署机器192.168.0.112

### 1.consul的介绍

### 2.consul的默认端口说明

- 8500:http端口，用于http接口和web ui访问
- 8300:service rpc端口，同一数据中心consul server之间通过该端口进行通信
- 8301:serf lan端口，同一数据中心不同consul client通过该端口通信，用于处理当前datacenter中LAN的gossip通信
- 8302:serf wan端口,不同数据中心consul server通过该端口通信，agent Server使用，处理与其他datacenter的gossip通信
- 8600:dns端口,用于已注册的服务发现

### 3.部署consul集群

参数说明:

```
agent: 表示启动 agent 进程
server: 表示 consul 为 server 模式
client: 表示 consul 为 client 模式
bootstrap: 表示这个节点是 Server-Leader
ui: 启动 Web UI, 默认端口 8500
node: 指定节点名称, 集群中节点名称唯一
client: 绑定客户端接口地址, 0.0.0.0 表示所有地址都可以访问
```

### 3.1启动第一个节点

```
docker run --name consul1 -d -p 8500:8500 -p 8300:8300 -p 8301:8301 -p 8302:8302 -p 8600:8600 consul agent -server -bootstrap-expect  2 -ui -bind=0.0.0.0 -client=0.0.0.0
```

#### 3.2启动第二个server节点，并加入consul1

##### 3.2.1 查看第一个server节点的ip地址（注意:这种方式下容器重启后ip可能发生变化）

```
docker inspect --format '{{ .NetworkSettings.IPAddress }}' consul1
172.17.0.5
```

##### 3.3.2 启动第二个server节点

```
docker run --name consul2 -d -p 8501:8500 consul agent -server -ui -bind=0.0.0.0 -client=0.0.0.0 -join 172.17.0.5
```

#### 3.2启动第三个server节点，并加入sonsul

```
docker run --name consul3 -d -p 8502:8500 consul agent -server -ui -bind=0.0.0.0 -client=0.0.0.0 -join 172.17.0.5
```

#### 3.3查询consul集群成员消息

```
docker exec -it consul1 consul members
```

#### 3.4 进入ui页面

http:192.168.0.112:8500

