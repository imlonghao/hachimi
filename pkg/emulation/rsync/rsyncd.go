package rsync

import (
	"bufio"
	"fmt"
	"hachimi/pkg/types"
	"io"
	"net"
	"strings"
)

// HandleRsync 处理rsync协议 27 版本 当前只能处理列出模块
func HandleRsync(conn net.Conn, session *types.Session) {
	//status := 0
	session.Service = "rsync"
	// 1. 握手：发送服务端协议版本
	_, err := fmt.Fprintf(conn, "@RSYNCD: %d\n", 27)
	if err != nil {
		return
	}
	rd := bufio.NewReader(conn)
	// 2. 读取客户端协议版本
	clientGreeting, err := rd.ReadString('\n')
	if err != nil {
		return
	}
	if !strings.HasPrefix(clientGreeting, "@RSYNCD: ") {
		return
	}

	for {
		// 3. 读取客户端请求的模块名
		requestedModule, err := rd.ReadString('\n')
		if err != nil {
			return
		}
		if strings.HasPrefix(requestedModule, "@RSYNCD: EXIT") {
			return
		}
		requestedModule = strings.TrimSpace(requestedModule)
		// 如果请求模块为空或请求列出模块，返回模块列表
		if requestedModule == "" || requestedModule == "#list" {
			_, err := io.WriteString(conn, formatModuleList())
			if err != nil {
				return
			}
			_, err = io.WriteString(conn, "@RSYNCD: EXIT\n")
			if err != nil {
				return
			}

			return

		}
		_, _ = io.WriteString(conn, "@RSYNCD: OK\n")
		//if requestedModule in modules
		if strings.HasSuffix(requestedModule, "//") {
			//status = 1
			sendFileEntry(&Conn{Writer: conn}, &file{name: "test.txt"}, nil, 0)

		}

	}

}

var modules = []string{"wwwdata", "backup", "data", "html"}

func formatModuleList() string {
	if len(modules) == 0 {
		return ""
	}
	var list strings.Builder
	for _, mod := range modules {
		fmt.Fprintf(&list, "%s\t%s\n",
			mod,
			mod)
	}
	return list.String()
}

type file struct {
	name  string
	size  int64
	mode  uint32
	mtime int64
	uid   int
	gid   int
}

func sendFileEntry(conn *Conn, f *file, last *file, flags uint16) error {
	// 初始化 flags
	flags = 0
	//TODO

	return nil
}

const (
	XMIT_TOP_DIR             = (1 << 0)
	XMIT_SAME_MODE           = (1 << 1)
	XMIT_EXTENDED_FLAGS      = (1 << 2)
	XMIT_SAME_RDEV_pre28     = XMIT_EXTENDED_FLAGS /* Only in protocols < 28 */
	XMIT_SAME_UID            = (1 << 3)
	XMIT_SAME_GID            = (1 << 4)
	XMIT_SAME_NAME           = (1 << 5)
	XMIT_LONG_NAME           = (1 << 6)
	XMIT_SAME_TIME           = (1 << 7)
	XMIT_SAME_RDEV_MAJOR     = (1 << 8)
	XMIT_HAS_IDEV_DATA       = (1 << 9)
	XMIT_SAME_DEV            = (1 << 10)
	XMIT_RDEV_MINOR_IS_SMALL = (1 << 11)
)
const (
	S_IFMT   = 0o0170000 // bits determining the file type
	S_IFDIR  = 0o0040000 // Directory
	S_IFCHR  = 0o0020000 // Character device
	S_IFBLK  = 0o0060000 // Block device
	S_IFREG  = 0o0100000 // Regular file
	S_IFIFO  = 0o0010000 // FIFO
	S_IFLNK  = 0o0120000 // Symbolic link
	S_IFSOCK = 0o0140000 // Socket
)
