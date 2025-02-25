package plugin

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"hachimi/pkg/types"
	"hachimi/pkg/utils"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

var TitleList []string
var ServeList []string
var TitleIndex int
var ServeIndex int

func nextTitle() string {
	title := TitleList[TitleIndex]
	TitleIndex = (TitleIndex + 1) % len(TitleList)
	return title
}
func nextServe() string {
	server := ServeList[ServeIndex]
	ServeIndex = (ServeIndex + 1) % len(ServeList)
	return server
}
func init() {
	TitleList = make([]string, 0)
	ServeList = make([]string, 0)
	//从文件按行加载Title
	file, err := os.Open("titles.txt")
	if err != nil {
		log.Fatalf("Failed opening file: %s", err)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		TitleList = append(TitleList, scanner.Text())
	}
	file.Close()
	//从文件按行加载Server
	file, err = os.Open("servers.txt")
	if err != nil {
		log.Fatalf("Failed opening file: %s", err)

	}
	scanner = bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		ServeList = append(ServeList, scanner.Text())
	}
	file.Close()
}
func extractString(input, start, end string) (string, error) {
	// 查找开始标记的位置
	startIndex := strings.Index(input, start)
	if startIndex == -1 {
		return "", fmt.Errorf("未找到开始标记")
	}

	// 跳过开始标记
	startIndex += len(start)

	// 查找结束标记的位置
	endIndex := strings.Index(input[startIndex:], end)
	if endIndex == -1 {
		return "", fmt.Errorf("未找到结束标记")
	}

	// 提取子字符串
	result := input[startIndex : startIndex+endIndex]

	return result, nil
}

func RequestHandler(plog *types.Http, ctx *fasthttp.RequestCtx) {
	ctx.SetConnectionClose()

	if strings.HasPrefix(string(ctx.URI().RequestURI()), "/_") {
		plog.Service = "Elasticsearch"
		esSearch(ctx)
		return
	}
	plog.Method = string(ctx.Method())
	plog.Host = string(ctx.Host())
	plog.UA = string(ctx.UserAgent())
	plog.Body = utils.EscapeBytes(ctx.Request.Body())
	plog.RawHeader = string(ctx.Request.Header.Header())
	ctx.Request.Header.VisitAll(func(key, value []byte) {
		plog.Header[string(key)] = string(value)
	})
	if string(ctx.Request.Header.Cookie("rememberMe")) != "" {
		ctx.Response.Header.Set("Set-Cookie", "rememberMe=deleteMe; Path=/; Max-Age=0;")
	}
	ctx.QueryArgs().VisitAll(func(key, value []byte) {
		plog.UriParam[string(key)] = string(value)
	})
	ctx.PostArgs().VisitAll(func(key, value []byte) {
		plog.BodyParam[string(key)] = string(value)
	})
	ctx.URI().DisablePathNormalizing = true
	plog.Path = string(ctx.URI().RequestURI())
	Hash := string(ctx.URI().Hash())
	if Hash != "" {
		plog.Path += "#" + Hash

	}
	if ctx.Request.Header.Peek("Upgrade") != nil {
		if string(ctx.Request.Header.Peek("Upgrade")) == "websocket" {
			//ctx.Response.SetStatusCode(101)
			//ctx.Response.Header.Set("Upgrade", "websocket")
			//ctx.Response.Header.Set("Connection", "Upgrade")
			key := string(ctx.Request.Header.Peek("Sec-WebSocket-Key"))
			accept := CalcWebSocketAccept(key)
			//ctx.Response.Header.Set("Sec-WebSocket-Accept", accept)
			//ctx.Response.Header.Set("Sec-WebSocket-Version", "13")
			// 不要使用 Hijack 接管连接
			conn := ctx.Conn()
			buf := ""
			//conn.Write([]byte("HTTP/1.1 101 Switching Protocols\r\n"))
			//conn.Write([]byte("Upgrade: websocket\r\n"))
			//conn.Write([]byte("Connection: Upgrade\r\n"))
			//conn.Write([]byte("Sec-WebSocket-Accept: " + accept + "\r\n"))
			//conn.Write([]byte("Sec-WebSocket-Version: 13\r\n"))
			//conn.Write([]byte("\r\n"))
			buf += "HTTP/1.1 101 Switching Protocols\r\n"
			buf += "Upgrade: websocket\r\n"
			buf += "Connection: Upgrade\r\n"
			buf += "Sec-WebSocket-Accept: " + accept + "\r\n"
			buf += "Sec-WebSocket-Version: 13\r\n"
			buf += "\r\n"
			//因为有的客户端不HTTP头支持碎片化TCP数据包，所以这一次性发送
			conn.Write([]byte(buf))

			//io.Copy(io.Discard, conn)
			plog.Body = ""

			for {
				// 读取并解析 WebSocket 帧
				message, isClosed, err := readWebSocketFrame(conn)
				if err != nil {
					if err == io.EOF {
						//fmt.Println("Connection closed")
					} else {
						//fmt.Println("Error reading frame:", err)
					}
					return
				}
				plog.Body += utils.EscapeBytes(message)
				plog.Body += "-WS_DATA_END-" //TODO
				// 如果客户端发送了关闭帧，退出处理
				if isClosed {
					return
				}
			}
		} else if string(ctx.Request.Header.Peek("Upgrade")) == "IF-T/TLS 1.0" {
			HandleIFTTLS(plog, ctx)
			return

		}

	}
	if strings.HasPrefix(plog.Path, "/dana-na") {
		ctx.Response.Header.Set("Content-Type", "text/html")
		ctx.Response.Header.Set("Server", "nginx/1.18.0 (Ubuntu)")
		ctx.WriteString("<html><body>\n<PARAM NAME=\"ProductName\"     VALUE=\"PulseSecure_Host_Checker\">\n<PARAM NAME=\"ProductVersion\"  VALUE=\"22.7.1.907\">\n<PARAM NAME=\"DSSETUP_ALL_PARAMS\"     VALUE=\"DSSETUP_BUILD_VERSION;DSSETUP_DOWNLOAD_URL;DSSETUP_INSTALL_ROOT;DSSETUP_STARTUP_EXE;DSSETUP_UNINSTALL_EXE;DSSETUP_CLIENT_VERSION_PULSE;DSSETUP_CLIENT_VERSION;DSSETUP_SECURITY_PATCH;DSSETUP_REDIST;DSSETUP_COMP_MSVCR80;DSSETUP_COMP_MSVCP80;DSSETUP_COMP_MSATL80;DSSETUP_COMP_MSMFC80;BrowserType;BrowserProxy;BrowserProxySettings;DSIDCookie;DSPREAUTHCookie;ProductName;ProductVersion;DownloadPath;DownloadPath64;locale;enable_logging;DisplayName;PreviousNames;StartSessionReDir;UninstallReDir;ReDirURL;DisableCancel;AutoInstall;AdminPrivilegeRequired;InstallationRoot;IniFilePath;AutoStart;upgradeMode;UninstallSync;LaunchInstallerOnly;launchFromApplication;WaitStartTimeout;WaitEndTimeout;hotfix;CompatibleVersion;uploadlog;meeting_type;PleaseWaitTimeout;DSSETUP_DISABLE_LOGGING;DSSETUP_PROMPT_FOR_DOWNLOAD;DSSETUP_DEPENDENCIES;DSSETUP_REDIST_EXCLUDED;ProductManifest;DSSETUP_APP_VERSION_MIN;DSSETUP_APP_VERSION_MAX;SignInId;DSSETUP_AX_REVISION;DSSETUP_INSTALLER_PARAMETER;InstallerType\">\n<PARAM NAME=\"DSSETUP_BUILD_VERSION\"  VALUE=\"22.7.1.907\">\n<PARAM NAME=\"DSSETUP_DOWNLOAD_URL\"   VALUE=\"/dana-cached/sc/PSSetupClientInstaller.exe\">\n<PARAM NAME=\"DSSETUP_INSTALL_ROOT\"   VALUE=\"AppData\">\n<PARAM NAME=\"DSSETUP_STARTUP_EXE\"    VALUE=\"PulseSetupClient.exe\">\n<PARAM NAME=\"DSSETUP_UNINSTALL_EXE\"  VALUE=\"uninstall.exe\">\n<PARAM NAME=\"DSSETUP_CLIENT_VERSION_PULSE\" VALUE=\"22.7.1.907\">\n<PARAM NAME=\"DSSETUP_CLIENT_VERSION\" VALUE=\"8.1.102.61663\">\n<PARAM NAME=\"DSSETUP_SECURITY_PATCH\" VALUE=\"1\">\n<PARAM NAME=\"DSSETUP_REDIST\"         VALUE=\"DSSETUP_COMP_MSVCR80\"></body></html>")
		return
	}

	if strings.HasPrefix(plog.Path, "/manager/html") {
		tomcatManger(ctx)
		return
	}

	if portHandler(plog, ctx) {
		return
	}
	if string(ctx.Method()) == "HEAD" {
		return
	}

	r := router.New()
	r.GET("/", index)
	r.GET("/v1.16/version", dockerVersion)
	r.GET("/cgi-bin/nas_sharing.cgi", dlinkNas)
	r.NotFound = notFound
	ctx.Response.Header.Set("Server", "nginx/1.18.0 (Ubuntu)")
	r.Handler(ctx)

}
func dlinkNas(ctx *fasthttp.RequestCtx) {
	//删除
}
func CalcWebSocketAccept(key string) string {
	const wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	h := sha1.New()
	h.Write([]byte(key + wsGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
func readWebSocketFrame(conn net.Conn) ([]byte, bool, error) {
	// 读取第一个两个字节：FIN/Opcode 和 Payload len
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, false, err
	}

	fin := (header[0] & 0x80) != 0      // FIN 位
	opcode := header[0] & 0x0F          // Opcode
	mask := (header[1] & 0x80) != 0     // Mask 位
	payloadLen := int(header[1] & 0x7F) // 初步 Payload 长度
	isCloseFrame := opcode == 0x8       // Opcode 0x8 表示关闭帧

	// 检查特殊帧（如关闭帧）
	if isCloseFrame {
		return nil, true, nil
	}

	// 处理扩展的 Payload 长度
	if payloadLen == 126 {
		// 2 字节的扩展长度
		extended := make([]byte, 2)
		if _, err := io.ReadFull(conn, extended); err != nil {
			return nil, false, err
		}
		payloadLen = int(binary.BigEndian.Uint16(extended))
	} else if payloadLen == 127 {
		// 8 字节的扩展长度
		extended := make([]byte, 8)
		if _, err := io.ReadFull(conn, extended); err != nil {
			return nil, false, err
		}
		payloadLen = int(binary.BigEndian.Uint64(extended))
	}

	// 读取 Mask Key（如果有）
	var maskKey []byte
	if mask {
		maskKey = make([]byte, 4)
		if _, err := io.ReadFull(conn, maskKey); err != nil {
			return nil, false, err
		}
	}

	// 读取 Payload 数据
	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(conn, payload); err != nil {
		return nil, false, err
	}

	// 如果存在 Mask，需要对数据解码
	if mask {
		for i := 0; i < payloadLen; i++ {
			payload[i] ^= maskKey[i%4]
		}
	}

	// 返回解析后的 Payload 数据
	if fin {
		return payload, false, nil
	} else {
		// FIN 未设置（分片帧），本示例不处理
		return nil, false, fmt.Errorf("fragmented frames not supported")
	}
}
func portHandler(plog *types.Http, ctx *fasthttp.RequestCtx) bool {
	isHandle := true
	switch ctx.LocalAddr().(*net.TCPAddr).Port {
	//Content-Type: application/json; charset=UTF-8
	case 9200:
		plog.Service = "Elasticsearch"
		esSearch(ctx)
		return true
	case 2375: //docker api
		plog.Service = "docker-api"

		if strings.HasSuffix(string(ctx.Path()), "/version") {
			dockerVersion(ctx)
			return true
		}
		ctx.Response.SetStatusCode(404)
		ctx.WriteString("{\"message\":\"page not found\"}\n") //有一个换行符
		return true
	default:
		isHandle = false
	}

	return isHandle
}

func setHeaders(ctx *fasthttp.RequestCtx, headers string) {
	lines := strings.Split(headers, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			ctx.Response.Header.Set(parts[0], parts[1])
		}
	}
}
func dockerVersion(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.Header.Set("Server", "Docker/20.10.0 (linux)")
	ctx.Response.Header.Set("Docker-Experimental", "false")
	ctx.Response.Header.Set("Ostype", "linux")
	ctx.Response.Header.Set("Api-Version", "1.41")
	ctx.WriteString(`{"Platform":{"Name":"Docker Engine - Community"},"Components":[{"Name":"Engine","Version":"20.10.0","Details":{"ApiVersion":"1.41","Arch":"amd64","BuildTime":"2020-12-08T18:56:55.000000000+00:00","Experimental":"false","GitCommit":"eeddea2","GoVersion":"go1.13.15","KernelVersion":"3.10.0-1160.99.1.el7.x86_64","MinAPIVersion":"1.12","Os":"linux"}},{"Name":"containerd","Version":"1.6.27","Details":{"GitCommit":"a1496014c916f9e62104b33d1bb5bd03b0858e59"}},{"Name":"runc","Version":"1.1.11","Details":{"GitCommit":"v1.1.11-0-g4bccb38"}},{"Name":"docker-init","Version":"0.19.0","Details":{"GitCommit":"de40ad0"}}],"Version":"20.10.0","ApiVersion":"1.41","MinAPIVersion":"1.12","GitCommit":"eeddea2","GoVersion":"go1.13.15","Os":"linux","Arch":"amd64","KernelVersion":"3.10.0-1160.99.1.el7.x86_64","BuildTime":"2020-12-08T18:56:55.000000000+00:00"}
`)
}

func esSearch(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("X-Elastic-Product", "Elasticsearch")
	ctx.Response.Header.Set("Warning", `299 Elasticsearch-7.17.8-120eabe1c8a0cb2ae87cffc109a5b65d213e9df1 "Elasticsearch built-in security features are not enabled. Without authentication, your cluster could be accessible to anyone. See https://www.elastic.co/guide/en/elasticsearch/reference/7.17/security-minimal-setup.html to enable security."`)
	ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
	switch string(ctx.Path()) {
	case "/":
		ctx.WriteString(`{
  "name" : "e944cd511328",
  "cluster_name" : "elasticsearch",
  "cluster_uuid" : "o6VYCbWMTiW9jXIna8s-FA",
  "version" : {
    "number" : "7.14.1",
    "build_flavor" : "default",
    "build_type" : "docker",
    "build_hash" : "66b55ebfa59c92c15db3f69a335d500018b3331e",
    "build_date" : "2021-08-26T09:01:05.390870785Z",
    "build_snapshot" : false,
    "lucene_version" : "8.9.0",
    "minimum_wire_compatibility_version" : "6.8.0",
    "minimum_index_compatibility_version" : "6.0.0-beta1"
  },
  "tagline" : "You Know, for Search"
}
`)
		return
	case "/_aliases":

		ctx.WriteString(`{"a":{"aliases":{}},"b":{"aliases":{}},"c7":{"aliases":{}},"bs":{"aliases":{}},"tse":{"aliases":{}},"esdfv":{"aliases":{}}}
`)
	case "/_cat/indices":
		ctx.Response.Header.Set("Content-Type", "text/plain; charset=UTF-8")
		ctx.WriteString(`green  open .geoip_databases o6VYCbWMTiW9jXIna8s-FA 5 1  5 0  1.2mb  1.2mb
green  open test 		  o6VYCbWMTiW9jXIna8s-FA 5 1  0 0    1kb    1kb
`)
	case "/_cluster/health":
		ctx.WriteString(`{"cluster_name":"elasticsearch","status":"green","timed_out":false,"number_of_nodes":1,"number_of_data_nodes":1,"active_primary_shards":5,"active_shards":5,"relocating_shards":0,"initializing_shards":0,"unassigned_shards":0,"delayed_unassigned_shards":0,"number_of_pending_tasks":0,"number_of_in_flight_fetch":0,"task_max_waiting_in_queue_millis":0,"active_shards_percent_as_number":100.0}
`)
		return
	}

	if strings.HasSuffix(string(ctx.Path()), "/_search") {
		ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
		ctx.WriteString(`{
	"took": 123,
	"timed_out": false,
	"_shards": {
		"total": 1,
		"successful": 1,
		"skipped": 0,
		"failed": 0
	},
	"hits": {
		"total": {
			"value": 1,
			"relation": "eq"
		},
		"max_score": 1.0,
		"hits": [{
			"_index": "hahah-tset",
			"_type": "_doc",
			"_id": "1",
			"_score": 1.0,
			"_ignored": ["message.keyword"],
			"_source": {
				"message": "When you gaze into the abyss, the abyss also gazes into you."
			}
		}]
	}
}`)

	} else {

		ctx.WriteString(`{"error":{"root_cause":[{"type":"index_not_found_exception","reason":"no such index [5status]","resource.type":"index_or_alias","resource.id":"5status","index_uuid":"_na_","index":"5status"}],"type":"index_not_found_exception","reason":"no such index [test]","resource.type":"index_or_alias","resource.id":"5status","index_uuid":"_na_","index":"5status"},"status":404}`)
	}

}

func index(ctx *fasthttp.RequestCtx) {
	notFound(ctx)
}

func notFound(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Server", nextServe())
	//setHeaders(ctx, headers)
	ctx.SetContentType("text/html; charset=utf-8")
	ctx.SetBodyString(strings.ReplaceAll(data2, "#{title}", nextTitle()))
}

// HandleIFTTLS CVE-2025-0282
func HandleIFTTLS(plog *types.Http, ctx *fasthttp.RequestCtx) {
	conn := ctx.Conn()
	//conn.Write([]byte("HTTP/1.1 101 Switching Protocols\r\n"))
	//conn.Write([]byte("Content-type: application/octet-stream\r\n"))
	//conn.Write([]byte("Pragma: no-cache\r\n"))
	//conn.Write([]byte("Upgrade: IF-T/TLS 1.0\r\n"))
	//conn.Write([]byte("Connection: Upgrade\r\n"))
	//conn.Write([]byte("Sec-WebSocket-Version: 13\r\n"))
	//conn.Write([]byte("\r\n"))
	//因为有的客户端不HTTP头支持碎片化TCP数据包，所以这一次性发送
	buf := "HTTP/1.1 101 Switching Protocols\r\n"
	buf += "Content-type: application/octet-stream\r\n"
	buf += "Pragma: no-cache\r\n"
	buf += "Upgrade: IF-T/TLS 1.0\r\n"
	buf += "Connection: Upgrade\r\n"
	buf += "Sec-WebSocket-Version: 13\r\n"
	buf += "\r\n"
	conn.Write([]byte(buf))
	type Packet struct {
		Vendor  uint32
		Type    uint32
		Length  uint32
		SeqNo   uint32
		Payload []byte
	}
	header := make([]byte, 16) // Fixed-size header (4 fields of 4 bytes each)
	if _, err := io.ReadFull(conn, header); err != nil {
		//fmt.Println("Error reading header:", err)
		return
	}
	var Init bool
	for {
		var vendor, ptype, length, seqNo uint32
		buffer := bytes.NewReader(header)
		if err := binary.Read(buffer, binary.BigEndian, &vendor); err != nil {
			//fmt.Println("Error reading vendor:", err)
			break
		}
		if err := binary.Read(buffer, binary.BigEndian, &ptype); err != nil {
			//fmt.Println("Error reading type:", err)
			break
		}
		if err := binary.Read(buffer, binary.BigEndian, &length); err != nil {
			//fmt.Println("Error reading length:", err)
			break
		}
		if err := binary.Read(buffer, binary.BigEndian, &seqNo); err != nil {
			//fmt.Println("Error reading seqNo:", err)
			break
		}
		if length < 16 {
			//fmt.Println("Invalid length:", length)
			break

		}
		data := make([]byte, length-16)
		var n int
		var err error
		if n, err = io.ReadFull(conn, data); err != nil {
			if n > 0 {
				plog.Body += utils.EscapeBytes(data[:n])
			}

			//fmt.Println("Error reading data:", err)
			break
		}
		plog.Body += utils.EscapeBytes(data[:n])
		//fmt.Printf("Vendor: %d, Type: %d, Length: %d, SeqNo: %d, Payload: %s\n", vendor, ptype, length, seqNo, data)
		if !Init {

			conn.Write([]byte("\x00\x00\x55\x97\x00\x00\x00\x02\x00\x00\x00\x14\x00\x00\x01\xf5\x00\x00\x00\x02"))
			conn.Write([]byte("\x00\x00\x55\x97\x00\x00\x00\x05\x00\x00\x00\x14\x00\x00\x01\xf6\x00\x0a\x4c\x01"))
			Init = true
		}
	}
	// Parse the header
	io.Copy(io.Discard, conn)
	//plog.Body = ""
}

func tomcatManger(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Content-Type", "text/html;charset=utf-8")
	ctx.Response.Header.Set("Cache-Control", "private")
	if string(ctx.Request.Header.Peek("Authorization")) != "" {
		ctx.Response.Header.Set("Set-Cookie", "JSESSIONID=7CA0511239CF843052119408A234494F; Path=/manager; HttpOnly")
		ctx.Response.Header.Set("Content-Type", "text/html;charset=utf-8")
		ctx.WriteString(`<html>
<head>
<link rel="stylesheet" href="/manager/css/manager.css">
<title>/manager</title>
</head>

<body bgcolor="#FFFFFF">

<table cellspacing="4" border="0">
 <tr>
  <td colspan="2">
   <a href="https://tomcat.apache.org/" rel="noopener noreferrer">
    <img class=tomcat-logo alt="The Tomcat Servlet/JSP Container"
         src="/manager/images/tomcat.svg">
   </a>
   <a href="https://www.apache.org/" rel="noopener noreferrer">
    <img border="0" alt="The Apache Software Foundation" align="right"
         src="/manager/images/asf-logo.svg" style="width: 266px; height: 83px;">
   </a>
  </td>
 </tr>
</table>
<hr size="1" noshade="noshade">
<table cellspacing="4" border="0">
 <tr>
  <td class="page-title" bordercolor="#000000" align="left" nowrap>
   <font size="+2">Tomcat Web Application Manager</font>
  </td>
 </tr>
</table>
<br>

<table border="1" cellspacing="0" cellpadding="3">
 <tr>
  <td class="row-left" width="10%"><small><strong>Message:</strong></small>&nbsp;</td>
  <td class="row-left"><pre>OK</pre></td>
 </tr>
</table>
<br>

<table border="1" cellspacing="0" cellpadding="3">
<tr>
 <td colspan="4" class="title">Manager</td>
</tr>
 <tr>
  <td class="row-left"><a href="/manager/html/list;jsessionid=7CA0511239CF843052119408A234494F?org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">List Applications</a></td>
  <td class="row-center"><a href="/manager/../docs/html-manager-howto.html" rel="noopener noreferrer">HTML Manager Help</a></td>
  <td class="row-center"><a href="/manager/../docs/manager-howto.html" rel="noopener noreferrer">Manager Help</a></td>
  <td class="row-right"><a href="/manager/status;jsessionid=7CA0511239CF843052119408A234494F?org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">Server Status</a></td>
 </tr>
</table>
<br>

<table border="1" cellspacing="0" cellpadding="3">
<tr>
 <td colspan="6" class="title">Applications</td>
</tr>
<tr>
 <td class="header-left"><small>Path</small></td>
 <td class="header-left"><small>Version</small></td>
 <td class="header-center"><small>Display Name</small></td>
 <td class="header-center"><small>Running</small></td>
 <td class="header-left"><small>Sessions</small></td>
 <td class="header-left"><small>Commands</small></td>
</tr>
<tr>
 <td class="row-left" bgcolor="#FFFFFF" rowspan="2"><small><a href="/" rel="noopener noreferrer">&#47;</a></small></td>
 <td class="row-left" bgcolor="#FFFFFF" rowspan="2"><small><i>None specified</i></small></td>
 <td class="row-left" bgcolor="#FFFFFF" rowspan="2"><small>Welcome to Tomcat</small></td>
 <td class="row-center" bgcolor="#FFFFFF" rowspan="2"><small>true</small></td>
 <td class="row-center" bgcolor="#FFFFFF" rowspan="2"><small><a href="&#47;manager&#47;html&#47;sessions;jsessionid=7CA0511239CF843052119408A234494F?path=&#47;&amp;org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">0</a></small></td>
 <td class="row-left" bgcolor="#FFFFFF">
  &nbsp;<small>Start</small>&nbsp;
  <form class="inline" method="POST" action="&#47;manager&#47;html&#47;stop;jsessionid=7CA0511239CF843052119408A234494F?path=&#47;&amp;org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">  <small><input type="submit" value="Stop"></small>  </form>
  <form class="inline" method="POST" action="&#47;manager&#47;html&#47;reload;jsessionid=7CA0511239CF843052119408A234494F?path=&#47;&amp;org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">  <small><input type="submit" value="Reload"></small>  </form>
  <form class="inline" method="POST" action="&#47;manager&#47;html&#47;undeploy;jsessionid=7CA0511239CF843052119408A234494F?path=&#47;&amp;org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">  &nbsp;&nbsp;<small><input type="submit" value="Undeploy"></small>  </form>
 </td>
 </tr><tr>
 <td class="row-left" bgcolor="#FFFFFF">
  <form method="POST" action="&#47;manager&#47;html&#47;expire;jsessionid=7CA0511239CF843052119408A234494F?path=&#47;&amp;org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">
  <small>
  &nbsp;<input type="submit" value="Expire sessions">&nbsp;with idle &ge;&nbsp;<input type="text" name="idle" size="5" value="30">&nbsp;minutes&nbsp;
  </small>
  </form>
 </td>
</tr>
<tr>
 <td class="row-left" bgcolor="#C3F3C3" rowspan="2"><small><a href="/docs/" rel="noopener noreferrer">&#47;docs</a></small></td>
 <td class="row-left" bgcolor="#C3F3C3" rowspan="2"><small><i>None specified</i></small></td>
 <td class="row-left" bgcolor="#C3F3C3" rowspan="2"><small>Tomcat Documentation</small></td>
 <td class="row-center" bgcolor="#C3F3C3" rowspan="2"><small>true</small></td>
 <td class="row-center" bgcolor="#C3F3C3" rowspan="2"><small><a href="&#47;manager&#47;html&#47;sessions;jsessionid=7CA0511239CF843052119408A234494F?path=&#47;docs&amp;org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">0</a></small></td>
 <td class="row-left" bgcolor="#C3F3C3">
  &nbsp;<small>Start</small>&nbsp;
  <form class="inline" method="POST" action="&#47;manager&#47;html&#47;stop;jsessionid=7CA0511239CF843052119408A234494F?path=&#47;docs&amp;org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">  <small><input type="submit" value="Stop"></small>  </form>
  <form class="inline" method="POST" action="&#47;manager&#47;html&#47;reload;jsessionid=7CA0511239CF843052119408A234494F?path=&#47;docs&amp;org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">  <small><input type="submit" value="Reload"></small>  </form>
  <form class="inline" method="POST" action="&#47;manager&#47;html&#47;undeploy;jsessionid=7CA0511239CF843052119408A234494F?path=&#47;docs&amp;org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">  &nbsp;&nbsp;<small><input type="submit" value="Undeploy"></small>  </form>
 </td>
 </tr><tr>
 <td class="row-left" bgcolor="#C3F3C3">
  <form method="POST" action="&#47;manager&#47;html&#47;expire;jsessionid=7CA0511239CF843052119408A234494F?path=&#47;docs&amp;org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">
  <small>
  &nbsp;<input type="submit" value="Expire sessions">&nbsp;with idle &ge;&nbsp;<input type="text" name="idle" size="5" value="30">&nbsp;minutes&nbsp;
  </small>
  </form>
 </td>
</tr>
<tr>
 <td class="row-left" bgcolor="#FFFFFF" rowspan="2"><small><a href="/examples/" rel="noopener noreferrer">&#47;examples</a></small></td>
 <td class="row-left" bgcolor="#FFFFFF" rowspan="2"><small><i>None specified</i></small></td>
 <td class="row-left" bgcolor="#FFFFFF" rowspan="2"><small>Servlet and JSP Examples</small></td>
 <td class="row-center" bgcolor="#FFFFFF" rowspan="2"><small>true</small></td>
 <td class="row-center" bgcolor="#FFFFFF" rowspan="2"><small><a href="&#47;manager&#47;html&#47;sessions;jsessionid=7CA0511239CF843052119408A234494F?path=&#47;examples&amp;org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">0</a></small></td>
 <td class="row-left" bgcolor="#FFFFFF">
  &nbsp;<small>Start</small>&nbsp;
  <form class="inline" method="POST" action="&#47;manager&#47;html&#47;stop;jsessionid=7CA0511239CF843052119408A234494F?path=&#47;examples&amp;org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">  <small><input type="submit" value="Stop"></small>  </form>
  <form class="inline" method="POST" action="&#47;manager&#47;html&#47;reload;jsessionid=7CA0511239CF843052119408A234494F?path=&#47;examples&amp;org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">  <small><input type="submit" value="Reload"></small>  </form>
  <form class="inline" method="POST" action="&#47;manager&#47;html&#47;undeploy;jsessionid=7CA0511239CF843052119408A234494F?path=&#47;examples&amp;org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">  &nbsp;&nbsp;<small><input type="submit" value="Undeploy"></small>  </form>
 </td>
 </tr><tr>
 <td class="row-left" bgcolor="#FFFFFF">
  <form method="POST" action="&#47;manager&#47;html&#47;expire;jsessionid=7CA0511239CF843052119408A234494F?path=&#47;examples&amp;org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">
  <small>
  &nbsp;<input type="submit" value="Expire sessions">&nbsp;with idle &ge;&nbsp;<input type="text" name="idle" size="5" value="30">&nbsp;minutes&nbsp;
  </small>
  </form>
 </td>
</tr>
<tr>
 <td class="row-left" bgcolor="#C3F3C3" rowspan="2"><small><a href="/host-manager/" rel="noopener noreferrer">&#47;host-manager</a></small></td>
 <td class="row-left" bgcolor="#C3F3C3" rowspan="2"><small><i>None specified</i></small></td>
 <td class="row-left" bgcolor="#C3F3C3" rowspan="2"><small>Tomcat Host Manager Application</small></td>
 <td class="row-center" bgcolor="#C3F3C3" rowspan="2"><small>true</small></td>
 <td class="row-center" bgcolor="#C3F3C3" rowspan="2"><small><a href="&#47;manager&#47;html&#47;sessions;jsessionid=7CA0511239CF843052119408A234494F?path=&#47;host-manager&amp;org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">0</a></small></td>
 <td class="row-left" bgcolor="#C3F3C3">
  &nbsp;<small>Start</small>&nbsp;
  <form class="inline" method="POST" action="&#47;manager&#47;html&#47;stop;jsessionid=7CA0511239CF843052119408A234494F?path=&#47;host-manager&amp;org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">  <small><input type="submit" value="Stop"></small>  </form>
  <form class="inline" method="POST" action="&#47;manager&#47;html&#47;reload;jsessionid=7CA0511239CF843052119408A234494F?path=&#47;host-manager&amp;org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">  <small><input type="submit" value="Reload"></small>  </form>
  <form class="inline" method="POST" action="&#47;manager&#47;html&#47;undeploy;jsessionid=7CA0511239CF843052119408A234494F?path=&#47;host-manager&amp;org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">  &nbsp;&nbsp;<small><input type="submit" value="Undeploy"></small>  </form>
 </td>
 </tr><tr>
 <td class="row-left" bgcolor="#C3F3C3">
  <form method="POST" action="&#47;manager&#47;html&#47;expire;jsessionid=7CA0511239CF843052119408A234494F?path=&#47;host-manager&amp;org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">
  <small>
  &nbsp;<input type="submit" value="Expire sessions">&nbsp;with idle &ge;&nbsp;<input type="text" name="idle" size="5" value="30">&nbsp;minutes&nbsp;
  </small>
  </form>
 </td>
</tr>
<tr>
 <td class="row-left" bgcolor="#FFFFFF" rowspan="2"><small><a href="/manager/" rel="noopener noreferrer">&#47;manager</a></small></td>
 <td class="row-left" bgcolor="#FFFFFF" rowspan="2"><small><i>None specified</i></small></td>
 <td class="row-left" bgcolor="#FFFFFF" rowspan="2"><small>Tomcat Manager Application</small></td>
 <td class="row-center" bgcolor="#FFFFFF" rowspan="2"><small>true</small></td>
 <td class="row-center" bgcolor="#FFFFFF" rowspan="2"><small><a href="&#47;manager&#47;html&#47;sessions;jsessionid=7CA0511239CF843052119408A234494F?path=&#47;manager&amp;org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">2</a></small></td>
 <td class="row-left" bgcolor="#FFFFFF">
  <small>
  &nbsp;Start&nbsp;
  &nbsp;Stop&nbsp;
  &nbsp;Reload&nbsp;
  &nbsp;Undeploy&nbsp;
  </small>
 </td>
</tr><tr>
 <td class="row-left" bgcolor="#FFFFFF">
  <form method="POST" action="&#47;manager&#47;html&#47;expire;jsessionid=7CA0511239CF843052119408A234494F?path=&#47;manager&amp;org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">
  <small>
  &nbsp;<input type="submit" value="Expire sessions">&nbsp;with idle &ge;&nbsp;<input type="text" name="idle" size="5" value="30">&nbsp;minutes&nbsp;
  </small>
  </form>
 </td>
</tr>
</table>
<br>
<table border="1" cellspacing="0" cellpadding="3">
<tr>
 <td colspan="2" class="title">Deploy</td>
</tr>
<tr>
 <td colspan="2" class="header-left"><small>Deploy directory or WAR file located on server</small></td>
</tr>
<tr>
 <td colspan="2">
<form method="post" action="/manager/html/deploy;jsessionid=7CA0511239CF843052119408A234494F?org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">
<table cellspacing="0" cellpadding="3">
<tr>
 <td class="row-right">
  <small>Context Path (required):</small>
 </td>
 <td class="row-left">
  <input type="text" name="deployPath" size="20">
 </td>
</tr>
<tr>
 <td class="row-right">
  <small>XML Configuration file path:</small>
 </td>
 <td class="row-left">
  <input type="text" name="deployConfig" size="20">
 </td>
</tr>
<tr>
 <td class="row-right">
  <small>WAR or Directory path:</small>
 </td>
 <td class="row-left">
  <input type="text" name="deployWar" size="40">
 </td>
</tr>
<tr>
 <td class="row-right">
  &nbsp;
 </td>
 <td class="row-left">
  <input type="submit" value="Deploy">
 </td>
</tr>
</table>
</form>
</td>
</tr>
<tr>
 <td colspan="2" class="header-left"><small>WAR file to deploy</small></td>
</tr>
<tr>
 <td colspan="2">
<form method="post" action="/manager/html/upload;jsessionid=7CA0511239CF843052119408A234494F?org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12" enctype="multipart/form-data">
<table cellspacing="0" cellpadding="3">
<tr>
 <td class="row-right">
  <small>Select WAR file to upload</small>
 </td>
 <td class="row-left">
  <input type="file" name="deployWar" size="40">
 </td>
</tr>
<tr>
 <td class="row-right">
  &nbsp;
 </td>
 <td class="row-left">
  <input type="submit" value="Deploy">
 </td>
</tr>
</table>
</form>
</td>
</tr>
</table>
<br>

<table border="1" cellspacing="0" cellpadding="3">
<tr>
 <td colspan="2" class="title">Configuration</td>
</tr>
<tr>
 <td colspan="2" class="header-left"><small>Re-read TLS configuration files</small></td>
</tr>
<tr>
 <td colspan="2">
<form method="post" action="/manager/html/sslReload;jsessionid=7CA0511239CF843052119408A234494F?org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">
<table cellspacing="0" cellpadding="3">
<tr>
 <td class="row-right">
  <small>TLS host name (optional)</small>
 </td>
 <td class="row-left">
  <input type="text" name="tlsHostName" size="20">
 </td>
</tr>
<tr>
 <td class="row-right">
  &nbsp;
 </td>
 <td class="row-left">
  <input type="submit" value="Re-read">
 </td>
</tr>
</table>
</form>
</td>
</tr>
</table>
<br><table border="1" cellspacing="0" cellpadding="3">
<tr>
 <td colspan="2" class="title">Diagnostics</td>
</tr>
<tr>
 <td colspan="2" class="header-left"><small>Check to see if a web application has caused a memory leak on stop, reload or undeploy</small></td>
</tr>
<tr>
 <td class="row-left">
  <form method="post" action="/manager/html/findleaks;jsessionid=7CA0511239CF843052119408A234494F?org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">
   <input type="submit" value="Find leaks">
  </form>
 </td>
 <td class="row-left">
  <small>This diagnostic check will trigger a full garbage collection. Use it with extreme caution on production systems.</small>
 </td>
</tr>
<tr>
 <td colspan="2" class="header-left"><small>TLS connector configuration diagnostics</small></td>
</tr>
<tr>
 <td class="row-left">
  <form method="post" action="/manager/html/sslConnectorCiphers;jsessionid=7CA0511239CF843052119408A234494F?org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">
   <input type="submit" value="Ciphers">
  </form>
 </td>
 <td class="row-left">
  <small>List the configured TLS virtual hosts and the ciphers for each.</small>
 </td>
</tr>
<tr>
 <td class="row-left">
  <form method="post" action="/manager/html/sslConnectorCerts;jsessionid=7CA0511239CF843052119408A234494F?org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">
   <input type="submit" value="Certificates">
  </form>
 </td>
 <td class="row-left">
  <small>List the configured TLS virtual hosts and the certificate chain for each.</small>
 </td>
</tr>
<tr>
 <td class="row-left">
  <form method="post" action="/manager/html/sslConnectorTrustedCerts;jsessionid=7CA0511239CF843052119408A234494F?org.apache.catalina.filters.CSRF_NONCE=9200181ECF80D49760124C8619DD6E12">
   <input type="submit" value="Trusted Certificates">
  </form>
 </td>
 <td class="row-left">
  <small>List the configured TLS virtual hosts and the trusted certificates for each.</small>
 </td>
</tr>
</table>
<br><table border="1" cellspacing="0" cellpadding="3">
<tr>
 <td colspan="8" class="title">Server Information</td>
</tr>
<tr>
 <td class="header-center"><small>Tomcat Version</small></td>
 <td class="header-center"><small>JVM Version</small></td>
 <td class="header-center"><small>JVM Vendor</small></td>
 <td class="header-center"><small>OS Name</small></td>
 <td class="header-center"><small>OS Version</small></td>
 <td class="header-center"><small>OS Architecture</small></td>
 <td class="header-center"><small>Hostname</small></td>
 <td class="header-center"><small>IP Address</small></td>
</tr>
<tr>
 <td class="row-center"><small>Apache Tomcat/8.5.90</small></td>
 <td class="row-center"><small>9.0.4+11</small></td>
 <td class="row-center"><small>Oracle Corporation</small></td>
 <td class="row-center"><small>Windows 10</small></td>
 <td class="row-center"><small>10.0</small></td>
 <td class="row-center"><small>amd64</small></td>
 <td class="row-center"><small>AD-DC</small></td>
 <td class="row-center"><small>10.25.125.54</small></td>
</tr>
</table>
<br>

<hr size="1" noshade="noshade">
<center><font size="-1" color="#525D76">
 <em>Copyright &copy; 1999-2023, Apache Software Foundation</em></font></center>

</body>
</html>`)
		return
	}
	ctx.Response.Header.Set("WWW-Authenticate", "Basic realm=\"Tomcat Manager Application\"")
	ctx.Response.SetStatusCode(401)
	ctx.Response.Header.Set("Content-Type", "text/html;charset=ISO-8859-1")
	ctx.WriteString(`<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd">
<html>
 <head>
  <title>401 Unauthorized</title>
  <style type="text/css">
    <!--
    BODY {font-family:Tahoma,Arial,sans-serif;color:black;background-color:white;font-size:12px;}
    H1 {font-family:Tahoma,Arial,sans-serif;color:white;background-color:#525D76;font-size:22px;}
    PRE, TT {border: 1px dotted #525D76}
    A {color : black;}A.name {color : black;}
    -->
  </style>
 </head>
 <body>
   <h1>401 Unauthorized</h1>
   <p>
    You are not authorized to view this page. If you have not changed
    any configuration files, please examine the file
    <tt>conf/tomcat-users.xml</tt> in your installation. That
    file must contain the credentials to let you use this webapp.
   </p>
   <p>
    For example, to add the <tt>manager-gui</tt> role to a user named
    <tt>tomcat</tt> with a password of <tt>s3cret</tt>, add the following to the
    config file listed above.
   </p>
<pre>
&lt;role rolename="manager-gui"/&gt;
&lt;user username="tomcat" password="s3cret" roles="manager-gui"/&gt;
</pre>
   <p>
    Note that for Tomcat 7 onwards, the roles required to use the manager
    application were changed from the single <tt>manager</tt> role to the
    following four roles. You will need to assign the role(s) required for
    the functionality you wish to access.
   </p>
    <ul>
      <li><tt>manager-gui</tt> - allows access to the HTML GUI and the status
          pages</li>
      <li><tt>manager-script</tt> - allows access to the text interface and the
          status pages</li>
      <li><tt>manager-jmx</tt> - allows access to the JMX proxy and the status
          pages</li>
      <li><tt>manager-status</tt> - allows access to the status pages only</li>
    </ul>
   <p>
    The HTML interface is protected against CSRF but the text and JMX interfaces
    are not. To maintain the CSRF protection:
   </p>
   <ul>
    <li>Users with the <tt>manager-gui</tt> role should not be granted either
        the <tt>manager-script</tt> or <tt>manager-jmx</tt> roles.</li>
    <li>If the text or jmx interfaces are accessed through a browser (e.g. for
        testing since these interfaces are intended for tools not humans) then
        the browser must be closed afterwards to terminate the session.</li>
   </ul>
   <p>
    For more information - please see the
    <a href="/docs/manager-howto.html" rel="noopener noreferrer">Manager App How-To</a>.
   </p>
 </body>

</html>
`)
	return
}

const data2 = `<!DOCTYPE html>
<html>
<head>
<title>#{title}</title>
<style>
    body {
        width: 35em;
        margin: 0 auto;
        font-family: Tahoma, Verdana, Arial, sans-serif;
    }
</style>
</head>
<body>
<h1>Welcome to nginx!</h1>
<p>If you see this page, the nginx web server is successfully installed and
working. Further configuration is required.</p>

<p>For online documentation and support please refer to
<a href="http://nginx.org/">nginx.org</a>.<br/>
Commercial support is available at
<a href="http://nginx.com/">nginx.com</a>.</p>

<p><em>Thank you for using nginx.</em></p>
</body>
</html>
`
