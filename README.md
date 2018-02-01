# DDNS
使用CloudFlare API实现ddns

配置文件路径conf/conf.json，使用前请修改以下项：

- `x_auth_email`：CloudFlare的注册邮箱
- `x_auth_key`：CloudFlare的API Key
- `dns_record`：DNS记录
    + `type`：目前仅支持A记录
    + `name`：需要解析的域名
    + `content`：IP地址
    + `proxied`：CloudFlare CDN加速
        * `true`：启用
        * `false`：关闭

配置示例文件如下：

```
{
    "get_ip": {
        "url": [
            "http://pv.sohu.com/cityjson?ie=utf-8",
            "http://ip.taobao.com/service/getIpInfo.php?ip=myip"
        ],
        "retry": 3,
        "duration": 3000000000
    },
    "secret": {
        "x_auth_email": "mulin@bbcclive.com",
        "x_auth_key": "1234567890abcdef1234567890abcdef12345"
    },
    "dns_record": [
        {
            "type": "A",
            "name": "bbcclive.com",
            "content": "1.1.1.1",
            "proxied": true
        },
        {
            "type": "A",
            "name": "www.bbcclive.com",
            "content": "2.2.2.2",
            "proxied": true
        }
    ],
    "duration": 3000000000
}
```