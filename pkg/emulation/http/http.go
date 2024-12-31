package http

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"github.com/valyala/fasthttp"
	"hachimi/pkg/plugin"
	"hachimi/pkg/plugin/symbols"
	"hachimi/pkg/types"
	"log"
	"net"

	"os"
	"sync"
	"time"
)

var serverPool sync.Pool

func ServeHttp(c net.Conn, handler fasthttp.RequestHandler) error {
	v := serverPool.Get()
	if v == nil {
		v = &fasthttp.Server{}
	}
	s := v.(*fasthttp.Server)
	s.NoDefaultServerHeader = true
	s.NoDefaultContentType = true
	s.ReadBufferSize = 1024 * 1024 * 5
	s.DisableHeaderNamesNormalizing = true
	s.DisableKeepalive = true
	s.Handler = handler
	err := s.ServeConn(c)
	s.Handler = nil
	serverPool.Put(v)
	return err
}

var RequestHandlerFunc func(*types.Http, *fasthttp.RequestCtx)

func init() {
	scriptFileName := "./httpserver.go"
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
		return
	}
	//defer watcher.Close()
	err = watcher.Add(scriptFileName)
	if err != nil {
		RequestHandlerFunc = plugin.RequestHandler
		return
	}
	loadScript(scriptFileName)

	if RequestHandlerFunc == nil {
		panic("requestHandlerFunc == nil")
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Printf("Event: %s\n", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					// 文件被写入，重新加载脚本
					log.Println("Reloading script...")
					time.Sleep(1 * time.Second)
					loadScript(scriptFileName)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error watching file:", err)
			}
		}
	}()

}
func loadScript(fileName string) {
	i := interp.New(interp.Options{})

	i.Use(stdlib.Symbols)
	i.Use(symbols.Symbols)
	// 从文件中读取脚本内容
	scriptContent, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println("Error reading script file:", err)
		return
	}
	_, err = i.Eval(string(scriptContent))
	if err != nil {
		fmt.Println("Error evaluating script:", err)
		return
	}

	// 获取 requestHandler 函数
	requestHandlerValue, err := i.Eval("plugin.RequestHandler")
	if err != nil {
		fmt.Println("Error getting requestHandler:", err)
		return
	}

	// 将值转换为函数
	ok := false
	lrequestHandlerFunc, ok := requestHandlerValue.Interface().(func(*types.Http, *fasthttp.RequestCtx))
	if !ok {
		fmt.Println("Cannot convert value to function")
		return
	}

	// 更新 requestHandlerFunc
	RequestHandlerFunc = lrequestHandlerFunc
	fmt.Println("Script reloaded successfully")
}
