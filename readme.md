
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

