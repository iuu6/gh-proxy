# GH-PROXY
为解决GitHub访问问题的 基于Golang的GitHub反向代理下载
# 代码架构
代码使用Golang语言编写，附加json配置文件以及html前端文件
## Go语言
实现大部分功能并监听8080端口用于服务，根目录下的ico和html文件作为前台文件，json作为配置文件
## json配置文件
这是一个json配置文件的示例

```
{
    "white_list": ["iuu6"],
    "black_list": ["a/a","a"],
    "size_limit": 1073741824
}
```
它可以配置**特定用户/用户的特定仓库**的黑白名单，设置**最大文件下载大小**等1073741824就是1GB
**注意这只是一个参考示例**

# systemd进程守护配置

**注意**：这只是一个参考！
```
[Unit]
Description=GitHub API
After=network.target

[Service]
Type=simple
WorkingDirectory=/usr/local/ghapi
ExecStart=/usr/local/ghapi/main
Restart=always

[Install]
WantedBy=multi-user.target
```
