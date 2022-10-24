# 115drive-webdav

[![GitHub Actions](https://img.shields.io/github/workflow/status/gaoyb7/115drive-webdav/CI)](https://github.com/gaoyb7/115drive-webdav/actions)
[![Release](https://img.shields.io/github/v/release/gaoyb7/115drive-webdav?display_name=tag)](https://github.com/gaoyb7/115drive-webdav/releases)
[![Downloads](https://img.shields.io/github/downloads/gaoyb7/115drive-webdav/total)](https://github.com/gaoyb7/115drive-webdav/releases)
[![Docker Image](https://img.shields.io/docker/pulls/gaoyb7/115drive-webdav)](https://hub.docker.com/r/gaoyb7/115drive-webdav)

115 网盘 WebDav 服务，可配合支持 WebDAV 协议的客户端 App 食用，如 [Infuse](https://firecore.com/infuse)、[nPlayer](https://nplayer.com) 

新项目 rclone 改版，对比 115drive-webdav 功能更强大，支持 WebDav 服务，本地磁盘挂载，文件批量下载到本地等功能。https://github.com/gaoyb7/rclone-release

## 下载
https://github.com/gaoyb7/115drive-webdav/releases

## 运行
需要获取 115 网盘 Cookie 信息，包括 UID、CID、SEID，网页版 Cookie 时效较短，建议抓包 App 请求获取 Cookie，iOS 系统可使用 [Stream](https://apps.apple.com/cn/app/stream/id1312141691) 抓包，安卓系统使用抓包精灵
```bash
./115drive-webdav --host=0.0.0.0 --port=8080 --user=user --pwd=123456 --uid=xxxxxx --cid=xxxxxxx --seid=xxxxx
```
服务启动成功后，用支持 WebDav 协议的客户端连接即可，不支持浏览器直接打开

## Docker 运行
```bash
# 通过命令参数获取配置
docker run -d \
        -p 8081:8081 \
	--restart unless-stopped \
        gaoyb7/115drive-webdav:latest \
	--host=0.0.0.0 --port=8081 \
	--user=user --pwd=123456 \
	--uid=xxxxxx \
	--cid=xxxxxx \
	--seid=xxxxxx
	
# 通过配置文件获取配置
# /path/to/your/config 替换为实际配置文件地址
docker run -d \
        -p 8081:8081 \
	-v /path/to/your/config:/etc/115drive-webdav.json \
        --restart unless-stopped \
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
* 使用 Stream 抓包，参考 https://cloud.tencent.com/developer/article/1670286

### Android
* 方法一：可使用抓包精灵，类似 Stream。参考 https://play.google.com/store/apps/details?id=com.minhui.networkcapture&hl=zh&gl=US
* 方法二：使用 Charles 抓包，参考 https://myoule.zhipin.com/articles/c27b2972802dc15fqxB72Ny9Eg~~.html
