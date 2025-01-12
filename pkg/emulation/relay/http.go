package relay

/*
HTTP Relay
实现HTTP 的中间人代理 深度包分析
*/

import (
	"crypto/tls"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	proxy "github.com/yeqown/fasthttp-reverse-proxy/v2"
	"hachimi/pkg/config"
	"hachimi/pkg/types"
	"hachimi/pkg/utils"
	"io"
	"net"
	"sync"
	"time"
)

var serverPool sync.Pool

func HandleHttpRelay(src net.Conn, session *types.Session, configMap map[string]string) bool {
	targetHost := configMap["targetHost"] //目标地址
	service := configMap["service"]       //服务
	if service == "" {
		service = "http_relay"
	}
	isTls := configMap["isTls"] //是否是tls

	var tlsConfig *tls.Config

	if isTls == "true" {
		host, _, _ := net.SplitHostPort(targetHost)
		//panic("tls not supported")
		tlsConfig = &tls.Config{
			ServerName:         host,
			InsecureSkipVerify: true,
			MaxVersion:         tls.VersionTLS13,
			MinVersion:         tls.VersionSSL30,
		}
	}

	httpLog := &types.Http{Session: *session}
	httpLog.StartTime = session.StartTime
	httpLog.ID = uuid.New().String()
	httpLog.SessionID = session.ID
	httpLog.Header = make(map[string]string)
	httpLog.BodyParam = make(map[string]string)
	httpLog.UriParam = make(map[string]string)
	httpLog.IsHandled = true
	httpLog.Service = service
	proxyServer, _ := proxy.NewReverseProxyWith(proxy.WithAddress(targetHost), proxy.WithTLSConfig(tlsConfig), proxy.WithTimeout(time.Duration(config.GetPotConfig().TimeOut)*time.Second))
	v := serverPool.Get()
	if v == nil {
		v = &fasthttp.Server{}
	}
	s := v.(*fasthttp.Server)
	s.NoDefaultServerHeader = true
	s.NoDefaultContentType = true
	s.ReadBufferSize = 1024 * 1024 * 5
	s.DisableHeaderNamesNormalizing = true
	s.DisableKeepalive = false
	httpLog.Service = service
	s.Handler = func(fasthttpCtx *fasthttp.RequestCtx) {
		// 在 requestHandlerFunc 中传递 ctx
		RequestHandler(httpLog, fasthttpCtx, proxyServer.ServeHTTP, configMap)
	}
	serverPool.Put(v)
	err := s.ServeConn(src)
	if err != nil {
		httpLog.IsHandled = false
		io.ReadAll(src) //出错继续读取
	}
	return true
}

func RequestHandler(plog *types.Http, ctx *fasthttp.RequestCtx, next func(*fasthttp.RequestCtx), configMap map[string]string) {
	ctx.SetConnectionClose()
	ctx.Request.Header.VisitAll(func(key, value []byte) {
		plog.Header[string(key)] = string(value)
	})
	if string(ctx.Request.Header.Cookie("rememberMe")) != "" {
		ctx.Response.Header.Set("Set-Cookie", "rememberMe=deleteMe; Path=/; Max-Age=0;")
	}

	ctx.URI().DisablePathNormalizing = true
	plog.Path = string(ctx.URI().RequestURI())
	Hash := string(ctx.URI().Hash())
	if Hash != "" {
		plog.Path += "#" + Hash

	}
	isTls := configMap["isTls"]
	if isTls == "true" {
		// HTTP 代理HTTPS
		// https://github.com/valyala/fasthttp/blob/ce283fb97c2e0c4801e68fd6c362a81a8a5c74b5/client.go#L1432C26-L1432C33
		// https://github.com/valyala/fasthttp/issues/841
		ctx.URI().SetScheme("https")
	}

	plog.Method = string(ctx.Method())
	plog.Host = string(ctx.Host())
	plog.UA = string(ctx.UserAgent())
	ctx.QueryArgs().VisitAll(func(key, value []byte) {
		plog.UriParam[string(key)] = string(value)
	})
	ctx.PostArgs().VisitAll(func(key, value []byte) {
		plog.BodyParam[string(key)] = string(value)
	})
	plog.Body = utils.EscapeBytes(ctx.Request.Body())
	plog.RawHeader = string(ctx.Request.Header.Header())
	next(ctx)
	plog.EndTime = time.Now()
	plog.Duration = int(plog.EndTime.Sub(plog.StartTime).Milliseconds())
	config.Logger.Log(plog)
}
