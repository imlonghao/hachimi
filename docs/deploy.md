# 系统架构
```
                 ┌─────┐  ┌─────┐  ┌─────┐                    
                 │ POT │  │ POT │  │ POT │  ...               
                 └──┬──┘  └──┬──┘  └──┬──┘                    
                    │        │ logs   │                       
                    │   ┌────▼────┐   │                       
┌──────────┐        └──►│   NSQ   │◄──┘                       
│          │            └────┬────┘                           
│  Redash  │          ┌──────┴───────┐                        
│          │          │              │                        
└──────────┘     ┌────▼───┐     ┌────▼───┐                    
     ▲           │ hunter ├──┬──┤ hunter │  ...              
     │           └────────┘  │  └────────┘                    
     │                       │                                
     │                ┌──────▼───────┐                        
     │                │              │                        
     └────────────────┤  Clickhouse  │                        
                      │              │                        
                      └──────────────┘                        
POT: 蜜罐节点
NSQ: 消息队列服务器 NSQD
hunter: 分析节点
Clickhouse: 数据库
Redash: 可视化分析平台
```
<!-- https://asciiflow.com/#/share/eJyrVspLzE1VslIKzSvJLMlJTVFIKUosz8xLVzBS0lHKSaxMLQJKVscolaUWFWfm58UoWRnpxChVAGlLc1MgqxIkYmkGZJWkVpQAOTFKCujg0ZSeR1MaUNEEUoUxQUxMHharmhQC%2FEPANB6Onp4eUYZNgTlgDYwxg1Rh4twNdToSMyc%2FvRhVlBRj0ANx2h7MAMVrNLZYwBUpCJ9P2wUx1C84EGLBo%2BkteMMCZhmSS9CchRSsaIGL31hko4NSUxKLM9CMxuHDLYR8S5L7McIYd6BDDJqCw3o0%2F%2BKN3wnEKsF0AETntE1oLs4ozStJLQIy52DEwhKEJLaMFQQ3FMXjeD06QwGWZfErwuMDnKGMJ%2FyJMAR7mkHPXESlGdxWkJRioEoIJhtMtERBwTknMzk7I7%2B0OJUoazAAOW7FYghJjseZ32OUapVqAcVYjIQ%3D) -->
# 系统要求
## 操作系统

蜜罐节点的完整功能仅在 Linux 系统上可用. 其他除数据库以外的组件没有操作系统要求

如果蜜罐节点在其他操作系统上尝试运行 透明代理可能将不可用. 这种情况仅支持单端口监听

## 组件

### 蜜罐节点

蜜罐节点占用资源极低，在 1 核 256M 内存的 VPS 上也可以轻松运行.

### 消息队列服务器

消息队列服务器使用 NSQ. 按照消费者和生产者的数量和消费速度决定服务器配置.

在小于10节点的蜜网下 最低可以使用 1 核 1G 内存的 VPS.

### 分析节点

分析节点作为消费者从消息队列服务器中获取数据，分析数据并将结果存储到数据库中

需要更多的内存和 CPU. 请按照节点数量和消费速度决定服务器配置.

### 数据库

建议使用clickhouse作为数据存储, 也可以使用其他数据库, 但是需要自行实现数据存储逻辑.

### 可视化分析

可视化分析使用 Redash, 请按照 Redash 官方文档安装.

### 网络要求

节点需要接受来自互联网的所有 TCP和UDP 流量，如果你的节点有防火墙，请确保已经放行全部协议和全部端口的入站流量。



# 节点安装

### iptables 设置


如果服务器上正在使用监听端口的服务 你正在通过公网端口管理蜜罐节点，比如SSH. 

请在蜜罐节点上的 iptables 规则，跳过你所正常使用的端口。下面的教程以保留端口范围65521-65535为例


修改默认ssh 端口为 65532 让端口在规则范围外 避免影响SSH正常连接
```bash
sudo sed -i 's/#Port 22/Port 65532/g' /etc/ssh/sshd_config
```

重启 ssh 服务让配置生效
```bash
sudo systemctl restart sshd
```

安装 iptables-services
```bash
sudo yum install -y iptables-services
```

放行所有的流量
```bash
sudo iptables -P INPUT ACCEPT
sudo iptables -P FORWARD ACCEPT
sudo iptables -P OUTPUT ACCEPT
```
清空所有规则
```bash
sudo iptables -F INPUT
sudo iptables -F FORWARD
sudo iptables -F OUTPUT
sudo iptables -F
sudo iptables -X
sudo iptables -Z
sudo iptables -t mangle -F
```
创建 DIVERT 链
```bash
sudo  iptables -t mangle -N DIVERT
```
设置 DIVERT 链的规则 用于出站流量bypass
```bash
sudo  iptables -t mangle -A DIVERT -j MARK --set-mark 1
sudo  iptables -t mangle -A DIVERT -j ACCEPT
sudo  iptables -t mangle -I PREROUTING -p tcp -m socket -j DIVERT
sudo  iptables -t mangle -I PREROUTING -p udp -m socket -j DIVERT
```
设置 DIVERT 链的规则 用于入站流量透明代理
`$(ip -o -4 route show to default | awk '{print $5}')` 获取默认路由网卡  使用其他网卡请手动替换为你的默认网卡

`$(hostname -I | awk '{print $1}')` 获取本机IP 其他情况请手动替换为你的本机IP 

`123456` 为蜜罐节点听端口

某些极老版本的 iptables 可能不支持端口范围 请注意规则是否添加成功
```bash
sudo  iptables -t mangle -A PREROUTING -i $(ip -o -4 route show to default | awk '{print $5}') -p tcp -d  $(hostname -I | awk '{print $1}')  --dport 0:12344 -j TPROXY --tproxy-mark 0x1/0x1 --on-ip   $(hostname -I | awk '{print $1}')  --on-port 12345
sudo  iptables -t mangle -A PREROUTING -i  $(ip -o -4 route show to default | awk '{print $5}') -p tcp -d  $(hostname -I | awk '{print $1}') --dport 12346:65520 -j TPROXY --tproxy-mark 0x1/0x1 --on-ip   $(hostname -I | awk '{print $1}')  --on-port 12345
sudo  iptables -t mangle -A PREROUTING -i  $(ip -o -4 route show to default | awk '{print $5}') -p udp -d  $(hostname -I | awk '{print $1}') --dport 0:12344 -j TPROXY --tproxy-mark 0x1/0x1 --on-ip   $(hostname -I | awk '{print $1}')  --on-port 12345
sudo  iptables -t mangle -A PREROUTING -i  $(ip -o -4 route show to default | awk '{print $5}') -p udp -d  $(hostname -I | awk '{print $1}') --dport 12346:65520 -j TPROXY --tproxy-mark 0x1/0x1 --on-ip   $(hostname -I | awk '{print $1}')  --on-port 12345
```
蜜罐同时支持IPV6网络 请使用ip6tables设置ipv6规则

`$(ip -o -6 route show to default | awk '{print $5}')` 获取IPV6默认路由网卡  使用其他网卡请手动替换为你的默认网卡

`$(hostname -I | awk '{print $2}')` 获取本机IPV6 其他网卡请手动替换为你的本机IP

保存规则 启动iptables

```bash
sudo service iptables save
sudo systemctl start iptables
sudo systemctl enable iptables
#ipv6
sudo service ip6tables save
sudo systemctl start ip6tables
sudo systemctl enable ip6tables
```

### 证书生成

蜜罐节点的TLS证书 推荐使用自己的证书 也可以不使用外部证书 每次启动都会生成一个新的临时证书
```bash
openssl genpkey -algorithm RSA -out private.key
openssl req -new -key private.key -out certificate.csr  -subj "/C=CN/ST=BeiJing/L=test/O=test/OU=test/CN=test"
openssl x509 -req -days 3650 -in certificate.csr -signkey private.key -out certificate.crt
```
### 启动蜜罐节点

蜜罐节点可以使用命令行参数和配置文件来设置配置  部分配置只有在配置文件中可以设置 请参考[配置文件](config.md)

不加参数启动蜜罐节点 会将系统日志和输出到标准错误 同时将蜜罐日志输出到标准输出
```bash
./honeypot -c config.yaml
```
# NSQ 部署
在github上下载最新的NSQ稳定版本 [仓库地址](https://github.com/nsqio/nsq/releases/)

## 单实例部署

http-address 为 NSQD 的 WEB 管理地址 无鉴权功能 建议绑定到本地地址 避免暴露在公网
tcp-address 为 NSQD 的 TCP 服务地址 用于生产者和消费者连接
```bash
./nsqd -http-address 127.0.0.1:11451 -tcp-address :1337
```
消费者处理不完的溢出数据会暂时存储在当前目录下

配置参考 [NSQD 配置](https://nsq.io/components/nsqd.html)
### 认证配置

#### secret 认证

你需要搭建一个认证服务 用于认证生产者和消费者的身份 
这里使用一个开源的认证服务 [nsq-auth](https://github.com/zhimiaox/nsq-auth)
```bash
./nsq-auth -a 127.0.0.1:1919 --secret hachimitsu
```
这样就可以启动一个简单的无权限划分的鉴权服务

之后在启动 nsqd 时使用 --auth-http-address 参数可以指定认证服务地址
```bash
./nsqd -http-address 127.0.0.1:11451 -tcp-address :1337 --auth-http-address 127.0.0.1:1919
```

# 测试消息队列

使用nsq 自带的nsq_tail工具测试消息队列是否正常工作

`-nsqd-tcp-address` 为`nsqd`消息队列服务器地址

`-producer-opt=auth_secret,secret` `可选` 为生产者认证凭据 与认证服务一致

`--topic hachimi` 为消息队列主题 在蜜罐节点上设置的主题对应

```bash 
nsq_tail -nsqd-tcp-address  127.0.0.1:1337 --topic hachimi -producer-opt=auth_secret,hachimitsu
```

访问蜜罐 蜜罐节点会将访问信息发送到消息队列中 nsq_tail 消费者会将消息逐行打印到控制台

正常情况下可以看到终端打印出json 格式的访问日志字符串