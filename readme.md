
### 1. this is a golang project
   this design for cos/ceph

### 2. 当前进展：
    第五章："云储存"之改造上传接口
    批量查询文件信息功能，已查询没问题
    上传文件的接口需要改造。返回的值需要修改下
    
    第六章：基于redis完成断点续传
    todo:合并上传功能
    toto:需要完成取消上传分块 和查看文件上传的整体状态这两个功能点的编写
    todo:需要编写文件分块上传的整体测试用例
    todo:需要补充文件断点上传的基本原理
    第七章:ceph
    ceph的开发已经完成。但是还没有用docker去部署ceph，也没有进行测试
    第八章：OSS
    使用的是阿里云的oss。环境可以用minio代替。
            TODO：可以增加一个功能，实现客户端直传OSS
    第九章：RbbitMQ
    基本概念：
        Exchange:消息交换机，决定消息按什么规则，路由到哪个队列
        Queue：消息载体，每个消息都会被投放到一个或者多个队列
        Binding：绑定，把exchange和queue按照路由规则绑定起来
        Routing Key:路由关键字，exchange根据这关键字来投递消息
        Channel:消息通道，客户端的每个链接建立多个channel
        Producer:生产者
        Consumer:消费者

    Exchange的工作模式：
        Fanout:广播模式，转发到所有绑定交换机的Queue
        Direct:类似单播，Routing Key和BindingKey完全匹配
        Topic:类似组播，转发到符合通配符匹配的Queue
        Headers:请求头与消息头匹配,才能接受消息
        
     data:2021-06-26 21:02:12 
        完成了单体架构版本的所有方法。
        基本功能已经通过测试。
        剩下分块上传和ceph上传尚未测试。
        还剩下ceph文件的下载功能
     
     
### 3. 文件的校验值计算

### 校验算法类型：
| 校验算法类型 | 校验值长度 | 校验值类型           | 安全级别 | 计算效率 | 应用场景                                     |
| ------------ | ---------- | -------------------- | -------- | -------- | -------------------------------------------- |
| CRC(32/64)   | 4/8个字节  | 常称为校验码         | 最弱     | 最高     | 传输数据的校验，客户端和文件端要传输一个文件 |
| MD5          | 16个字节   | 常称为hash值，散列值 | 居中     | 居中     | 文件的校验和数据的签名                       |
| SHA1         | 20个字节   | 常称为hash值，散列值 | 最高     | 最低     | 文件的校验，文件的唯一标识                   |

如果对安全性要求更高则可以使用sha256

### 4. 秒传的原理

​	场景：1.用户上传  2.离线下载  3.好友分享

​	关键点：

​	1.计算文件的hash值（MD5,SHA1值等）

​	2.用户文件的关联

### 5. 相同文件的处理
    1.允许不同用户同时上传同一个文件
    2.先上传完成的，先入库
    3.后上传的只更新用户文件表，并删除已上传的文件

### 6. 分块上传与断点上传
    分块上传：文件切成多块，独立传输，上传完成后完成合并
    断点续传：传输暂停或者异常中断后，可基于原来进度重传
    
    几点说明:
    小文件不建议分块上传
    可以并行的上传分块，并且可以无序传输
    分块传输可以极大提高传输效率
    减少传输失败后重试的流量和时间



### 7.Ceph

**Cpeh的特点：**

部署简单，开源，客户端支持多语言

可靠性高，性能高，分布式，可扩展性强

**Ceph的组件：**

- OSD:用于集群中所有数据与对象的储存：储存/复制/平衡/恢复数据等等
- Monitor:监控集群状态，维护cluster MAP表，保证集群数据一致性。
- MDS：保存文件系统服务的元数据(OBJ/Block不需要该服务)
- GW：提供Amazon S3和Swif兼容的RESTful API的getway服务

**AWS S3术语：**

Region：存储数据所在的地理区域。

Endpoint：存储服务入口，Web服务入口点的URL。

Bucket：储存桶是S3中储存的基本实体，由对象数据和元数据组成。

Key:健是储存桶中对象的唯一标识符，桶内的每个对象都只能有一个ke


y


### go 微服务框架 go-mirco

### 安装使用proto
安装：
使用:
protoc --proto_path=service/account/proto  --go_out=service/account/proto  --micro_out=service/account/proto  service/account/proto/user.proto
protoc --proto_path=service/dbproxy/proto --go_out=service/dbproxy/proto --micro_out=service/dbproxy/proto  service/dbproxy/proto/dbproxy.proto