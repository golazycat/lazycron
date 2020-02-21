# lazycron

![](https://img.shields.io/badge/golang-1.13-green)
![](https://img.shields.io/badge/etcd-3.3.9-blue)
![](https://img.shields.io/badge/mongodb-4.2.3-orange)

lazycron是一个使用`golang1.13`，基于`etcd 3.3.9`，`mongodb 4.2.3`的简易master-worker分布式cron表达式调度执行框架。

## 依赖

### etcd

由于是一个分布式软件，本项目使用高可用分布式kv储存服务来进行worker的注册、任务的派发、分布式锁的实现等功能。

因此没有etcd，本软件将会无法工作。

etcd是完全开源的，在使用lazycron之前，需要至少搭建到一个etcd服务，您可以去[etcd官网](https://etcd.io/)了解更多etcd的知识并下载运行。

### mongodb

lazycron使用mongodb来储存任务运行的日志，包括任务调度、运行的时间，任务的输出结果和错误信息等信息。

没有mongodb，您将无法通过lazycron查看任务的执行情况。

因此在使用lazycron之前需要搭建好mongodb服务，安装方法见[mongodb官网](https://www.mongodb.com/)。

## 使用Docker安装(推荐)

安装好上述依赖，即可体验lazycron了。

推荐使用docker直接安装运行lazycron master：

```text
$ docker run --name lazycron-master -p 8070:8070 -d golazycat/lazycron:1.0 master --conf /etc/master.json
```

这将配置文件设置在容器的`/etc/master.json`下，但是默认是没有这个文件的，需要您在本地编辑好配置文件后上传到容器中：

```text
$ vim master.json
...(编辑配置文件)
$ docker cp master.json lazycron-master:/etc/master.json
...
$ docker restart lazycron-master
```

这样容器就按照您自己的配置文件启动了，关于配置文件如何编辑，见[配置文件说明](#配置)。

随后，就可以通过访问`localhost:8070`来访问master管理后台了。

如果想运行lazycron worker，则可以运行：

```text
$ docker run --name lazycron-worker -d golazycat/lazycron:1.0 worker --conf /etc/worker.json
```

随后也需要把worker的配置文件上传到容器的`/etc/worker.json`中并重启：

```text
$ vim worker.json
...(编辑配置文件)
$ docker cp worker.json lazycron-worker:/etc/worker.json
...
$ docker restart lazycron-worker
```

注意worker和master的etcd和mongodb配置必须一致，这样它们才会在同一个集群中。

## 下载安装 

lazycron只支持mac和linux操作系统，在[release](https://github.com/golazycat/lazycron/releases/tag/v1.0)中查找选择适合您系统的二进制文件下载。

下载到本地并解压后，从终端进入目录，目录下有三个可执行文件：

- master: 运行lazycron master
- worker: 运行lazycron worker
- standalone: 运行单机版lazycron

通过`--conf`参数可以指定配置文件路径(standalone不可以指定)，不指定使用默认配置，默认配置见目录下的[master.json](master.json)和[worker.json](worker.json)。

例如，配置文件为`/etc/master.json`，则可以通过以下命令来运行lazycron master：

```text
$ ./master --conf /etc/master.json
```

worker的运行同理。

注意worker和master的etcd和mongodb配置必须一致，这样它们才会在同一个集群中。

## 通过源代码build安装

见[build教程](#build教程)。

## 配置

配置使用json文件的方式

worker和master有一些通用配置，见下表：

配置名称|类型|说明|默认值
---|---|---|---
etcd.endpoints|string数组|etcd节点url配置|\["localhost:2379"\]
etcd.dial_timeout|int|etcd连接超时时间，单位为秒|5
run.n_thread|int|程序启动多少个线程|CPU核数
mongodb.connect_url|string|mongodb连接url|"mongodb://localhost:27017"
mongodb.connect_timeout|int|mongodb连接超时时间，单位为秒|5
mongodb.write_batch_size|int|mongodb写入数据batch大小|100
log.error_path|string|错误日志输出文件路径，默认不输出到文件|""

下面的配置项是master独有的：

配置名称|类型|说明|默认值
---|---|---|---
http.addr|string|http api的监听地址|""
http.port|int|http api的监听端口|8070
http.read_timeout|int|http请求读超时时间，单位为秒|5
http.write_timeout|int|http请求写超时时间，单位为秒|5
http.static_path|string|http静态文件路径|"./static"

下面的配置项是worker独有的：

配置名称|类型|说明|默认值
---|---|---|---
log_job|bool|运行job时是否输出日志。<br>如果为true，任何job被运行时会输出日志。<br>可能会导致日志很长|true



## HTTP API

启动master后，可以通过`ip:8070`来访问管理后台，通过界面可以直接管理任务。

如果想开发自己的应用调用lazycron，可以使用lazycron的HTTP API。

lazycron的所有API的Method都是POST，参数通过RequestBody传输，结果通过ResponseBody返回。返回的数据是一个json字符串，所有请求的返回格式一致：

参数|类型|说明
---|---|---
errno|int|请求的错误码，0表示没有错误
message|string|请求返回的信息，如果请求成功固定为"ok"，失败则为错误信息
data|object|请求返回的数据，没有数据为null。具体类型根据请求而定，见下面请求表。

所有已知的错误码为：

错误码|说明
---|---
0|没有错误
1|Http参数错误
2|Json格式错误
3|任务管理器错误
4|任务日志器错误
5|worker管理器错误

下面是所有的请求说明：

请求url|参数|请求成功data类型|说明
---|---|---|---
/job/save|job: 新增的job json数据：<br>{<br>"name": "任务名称",<br>"command": "任务执行的命令",<br>"cron_expr": "任务的cron表达式"<br>}|如果是新增，为null;<br>如果是更新，为旧的job的json数据|保存一个任务。这个接口会让新的任务被其它worker收到，并根据cron表达式调度执行。
/job/del|name: 要删除的任务名称|如果删除成功，为删除的job数据;<br>如果删除失败，为null|删除一个任务。这个接口会让worker停止这个任务并不再执行。
/job/list|无|job列表|列出所有任务
/job/kill|name: 要kill的任务名称|null|让worker kill这个任务，这会让正在运行这个任务的worker终止运行任务。但是不同于删除，后续还是会依据cron表达式重新调度执行该任务。
/job/log|name:要查询的任务名称<br>skip: int，分页参数，跳过多少个记录<br>limit: int，分页参数，限制多少个数据|job log列表|列出某个任务的执行日志
/worker/list|无|worker列表|列出当前所有的健康节点

## build教程

如果想在机器上自己complie这个项目，首先需要拉取项目代码并进入项目路径：

```text
$ git clone https://github.com/golazycat/lazycron.git
$ cd lazycron
```

随后有两种方式build：

### 本机build

仅支持mac或linux系统。

需要机器安装有`go 1.13`环境，并且确保已经启用`go mod`功能。最好设置了国内的go代理。如果没有设置，通过以下命令设置好：

```text
$ go env -w GO111MODULE=on
$ go env -w GOPROXY=https://goproxy.cn,direct
```

直接运行[build.sh](build.sh)脚本即可：

```text
$ sh build.sh
```

build成功后，根目录下的`out`目录即为输出目录，下面有lazycron的所有可执行文件。

### 通过Docker build

支持所有系统，仅需要机器上安装好Docker。

通过以下命令build：

```text
$ docker build -t golazycat/lazycron:1.0 .
```

这会创建一个叫做`golazycat/lazycron:1.0`的镜像在本地，随后的运行方式就和[使用Docker安装(推荐)](#使用Docker安装(推荐))中的一样了。只不过不再需要从Docker Hub下载镜像了。
