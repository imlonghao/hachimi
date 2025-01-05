# hachimi

一个分布式蜜网系统，用于收集和分析来自互联网的背景噪音 (Internet Background Noise)。

互联网背景噪音是指来自互联网的大量扫描、攻击、恶意软件传播等流量。这些流量通常是由恶意软件、僵尸网络、漏洞扫描器等产生的，对网络安全和数据分析有很大的帮助。

该项目通过linux透明代理来实现全端口监听，根据请求数据推测请求协议并模拟对应服务的响应。
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

蜜罐节点将流量分析后发送到消息队列中，分析节点消费消息队列中的数据，将数据存储到数据库中，最后通过可视化工具展示数据。
```
<!-- https://asciiflow.com/#/share/eJyrVspLzE1VslIKzSvJLMlJTVFIKUosz8xLVzBS0lHKSaxMLQJKVscolaUWFWfm58UoWRnpxChVAGlLc1MgqxIkYmkGZJWkVpQAOTFKCujg0ZSeR1MaUNEEUoUxQUxMHharmhQC%2FEPANB6Onp4eUYZNgTlgDYwxg1Rh4twNdToSMyc%2FvRhVlBRj0ANx2h7MAMVrNLZYwBUpCJ9P2wUx1C84EGLBo%2BkteMMCZhmSS9CchRSsaIGL31hko4NSUxKLM9CMxuHDLYR8S5L7McIYd6BDDJqCw3o0%2F%2BKN3wnEKsF0AETntE1oLs4ozStJLQIy52DEwhKEJLaMFQQ3FMXjeD06QwGWZfErwuMDnKGMJ%2FyJMAR7mkHPXESlGdxWkJRioEoIJhtMtERBwTknMzk7I7%2B0OJUoazAAOW7FYghJjseZ32OUapVqAcVYjIQ%3D) -->

## 部署
[部署文档](docs/deploy.md)


## 数据展示

![demo](demo.png)
![img.png](demo1.png)

## 数据分析

## 开源数据
开源互联网背景噪音数据集，包含了2024年9月到2025年1月的约1000万条HTTP请求数据。，数据集为Parquet格式，包含了请求的时间、源IP端口、请求方法、请求路径、请求头、请求体等信息。
[数据集地址](https://huggingface.co/datasets/burpheart/Internet-background-noise)