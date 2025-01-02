package config

import "errors"

type ForwardingRule struct {
	// Port 匹配端口 0 为所有端口
	Port int `toml:"port"`
	// Handler 处理器 目前有 relay_http relay_tcp(null 为只记录流量)
	Handler string `toml:"handler"`
	// Config 配置 处理器的配置 目前只有 relay_http relay_tcp 需要配置
	Config map[string]string `toml:"config"`
}

/*
relay_tcp 配置:
service 服务名
targetAddr 上游主机名:端口
sendSession 是否发送session信息  启用会连接后发送一行session的序列化信息用于 传递真实ip等信息 上游需要自行处理

relay_http 配置:
service 服务名
targetHost 上游主机名:端口
isTls 是否启用tls
真实ip 地址会从xff头中传递
另外relay_http会深度解析 http 的请求和响应 relay_tcp只会在session 中记录数据
*/
func validateForwardingRule(rules *[]ForwardingRule) error {
	for _, rule := range *rules {
		if rule.Handler == "relay_http" {
			if rule.Config["targetHost"] == "" {
				return errors.New("relay_http rule targetHost is empty")
			}
		} else if rule.Handler == "relay_tcp" {
			if rule.Config["targetAddr"] == "" {
				return errors.New("relay_tcp rule targetAddr is empty")
			}
		} else {
			return errors.New("unknown handler " + rule.Handler)
		}
	}
	return nil

}
