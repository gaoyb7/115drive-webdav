# 115drive-webdav

[![GitHub Actions](https://img.shields.io/github/workflow/status/gaoyb7/115drive-webdav/CI)](https://github.com/gaoyb7/115drive-webdav/actions)
[![Release](https://img.shields.io/github/v/release/gaoyb7/115drive-webdav?display_name=tag)](https://github.com/gaoyb7/115drive-webdav/releases)
[![Downloads](https://img.shields.io/github/downloads/gaoyb7/115drive-webdav/total)](https://github.com/gaoyb7/115drive-webdav/releases)
[![Docker Image](https://img.shields.io/docker/pulls/gaoyb7/115drive-webdav)](https://hub.docker.com/r/gaoyb7/115drive-webdav)

115 网盘 WebDav 服务，可配合支持 WebDAV 协议的客户端 App 食用，如 [Infuse](https://firecore.com/infuse)、[nPlayer](https://nplayer.com) 

新项目 rclone 改版，对比 115drive-webdav 功能更强大，支持 WebDav 服务，本地磁盘挂载，文件批量下载到本地等功能，提供预编译包可自行下载试用 https://github.com/gaoyb7/rclone-release

## 下载
https://github.com/gaoyb7/115drive-webdav/releases

* 暂只支持 Linux x86_64 版本，程序下载后，需要额外下载加解密库 https://raw.githubusercontent.com/gaoyb7/115drive-webdav/main/115/libencode115.so 到 /usr/lib 目录下才可运行。
* 需要使用 App 的 Cookie，浏览器版本的 Cookie 无法正常下载文件，抓包方法参考下面说明。

## 运行
需要获取 115 网盘 Cookie 信息，包括 UID、CID、SEID，网页版 Cookie 时效较短，建议抓包 App 请求获取 Cookie，iOS 系统可使用 [Stream](https://apps.apple.com/cn/app/stream/id1312141691) 抓包
```bash
./115drive-webdav --host=0.0.0.0 --port=8080 --user=user --pwd=123456 --uid=xxxxxx --cid=xxxxxxx --seid=xxxxx
```
服务启动成功后，用支持 WebDav 协议的客户端连接即可，不支持浏览器直接打开

替换后缀请在config中替换 `replace`的内容 默认`"*==."`,多个`"a==b|c==d|x==y"`

## Docker 运行
```bash
# 通过命令参数获取配置
docker run -d -p 8081:8081 gaoyb7/115drive-webdav \
	--host=0.0.0.0 --port=8081 \
	--user=user --pwd=123456 \
	--uid=xxxxxx \
	--cid=xxxxxx \
	--seid=xxxxxx
	
# 通过配置文件获取配置
docker run -d -p 8081:8081 \
	-v /path/to/your/config:/etc/115drive-webdav.json \
	gaoyb7/115drive-webdav \
	--config /etc/115drive-webdav.json
```

## 参数说明
```bash
--host
    服务监听地址，默认 0.0.0.0
--port
    服务监听端口，默认 8080
--user
    WebDav 账户用户名，默认 user
--pwd
    WebDav 账户密码，默认 123456
--uid
    115 网盘 Cookie，UID
--cid
    115 网盘 Cookie，CID
--seid
    115 网盘 Cookie，SEID
--config
    从文件中读取配置，参考 config.json.example
```

## 功能支持

- [x] 文件/文件夹查看
- [x] 文件下载
- [x] WebDav 权限校验
- [x] WebDav 在线视频播放
- [ ] 文件上传
- [x] 文件重命名
- [x] 文件删除
- [x] 文件移动

## App Cookie 获取方法
### iOS
* 安装 [Stream](https://apps.apple.com/cn/app/stream/id1312141691)

* 打开 Stream，根据提示设置 HTTPS 抓包，安装证书

* 点击开始抓包

* 切换到 115 App，点开一个视频，开始播放

* 再次切换回 Stream，停止抓包，在抓包历史中，找到 115 相关的域名请求，获取到 Cookie 信息

<img src="https://user-images.githubusercontent.com/9281603/183956374-f3eb563b-3c04-4285-a0e8-af3eda13e42a.png" width="50%">

### Android
参考 https://myoule.zhipin.com/articles/c27b2972802dc15fqxB72Ny9Eg~~.html
