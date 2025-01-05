package relay

import (
	"encoding/json"
	"hachimi/pkg/types"
	"hachimi/pkg/utils"
	"io"
	"log"
	"net"
)

func HandleTCPRelay(src net.Conn, session *types.Session, config map[string]string) bool {
	targetAddr := config["targetAddr"]  //目标地址
	sendSesion := config["sendSession"] //是否发送session
	dst, err := net.Dial("tcp", targetAddr)
	if err != nil {
		log.Printf("Failed to connect to target server: %v", err)
		return false
	}
	defer dst.Close()
	//连接 两个连接
	if sendSesion == "true" {
		sess, _ := utils.ToMap(session)
		jsonData, _ := json.Marshal(sess)
		dst.Write(jsonData)
		dst.Write([]byte("\n"))
	}
	go func() {
		io.Copy(src, dst)
	}()
	io.Copy(dst, src)

	return true
}
