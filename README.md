# 115drive-webdav

115 网盘 WebDav 服务，可配合支持 WebDAV 协议的客户端 App 食用，如 [Infuse](https://firecore.com/infuse)、[nPlayer](https://nplayer.com) 

## 下载
https://github.com/gaoyb7/115drive-webdav/releases

## 运行
需要获取 115 网盘 Cookie 信息，包括 UID、CID、SEID，建议抓包 App 请求获取，iOS 系统可使用 [Stream](https://apps.apple.com/cn/app/stream/id1312141691) 抓包
```bash
./115drive-webdav --host=0.0.0.0 \
    --port=8080 \
    --user=user \
    --pwd=123456 \
    --uid=xxxxxx \
    --cid=xxxxxxx \
    --seid=xxxxx
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
```

## 功能支持

- [x] 文件/文件夹查看
- [x] 文件下载
- [x] WebDav 权限校验
- [x] WebDav 在线视频播放
- [ ] 文件上传
- [ ] 文件重命名
- [ ] 文件删除
- [ ] 文件移动
