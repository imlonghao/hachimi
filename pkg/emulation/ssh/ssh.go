package ssh

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh/terminal"
	"hachimi/pkg/config"
	"hachimi/pkg/types"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHSession struct {
	types.Session
	ID            string    `gorm:"primaryKey" json:"id"`
	SessionID     string    `gorm:"index" json:"session_id"`
	StartTime     time.Time `gorm:"index" json:"start_time"`
	EndTime       time.Time `gorm:"index" json:"end_time"`
	Duration      int       `json:"duration"`
	ClientVersion string    `json:"client_version"`
	Shell         string    `json:"shell"`
	Request       string    `json:"request"`
	Error         bool      `json:"error"`
	PublicKey     string    `json:"public_key"`
	Service       string    `json:"service"`
	User          string    `json:"user"`
	Data          string    `json:"data"`
	IsInteract    bool      `json:"is_interact"`
	PassWord      string    `json:"password"`
}

var (
	errBadPassword = errors.New("permission denied")
	ServerVersions = []string{
		"SSH-2.0-OpenSSH_8.4",
	}
)

func HandleSsh(conn net.Conn, session *types.Session) {
	var s SSHSession
	s.Session = *session
	serverConfig := &ssh.ServerConfig{
		MaxAuthTries:      6,
		PasswordCallback:  s.PasswordCallback,
		PublicKeyCallback: s.PublicKeyCallback,
		ServerVersion:     ServerVersions[0],
	}
	s.ID = uuid.New().String()
	s.Service = "ssh"
	s.SessionID = session.ID
	s.StartTime = time.Now()
	signer, _ := ssh.NewSignerFromSigner(config.SshPrivateKey)
	serverConfig.AddHostKey(signer)
	s.HandleConn(conn, serverConfig)
	s.EndTime = time.Now()
	s.Duration = int(s.EndTime.Sub(s.StartTime).Milliseconds())
	config.Logger.Log(&s)

}

// TableName 设置表名
func (s *SSHSession) TableName() string {
	return "ssh_logs"
}
func (s *SSHSession) ToMap() (map[string]interface{}, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func (s *SSHSession) PublicKeyCallback(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	s.User = conn.User()
	s.PublicKey = strings.Trim(strconv.Quote(string(key.Marshal())), `"`)

	//time.Sleep(100 * time.Millisecond)
	return nil, nil
}

func (s *SSHSession) PasswordCallback(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	s.User = conn.User()
	s.PassWord = strings.Trim(strconv.Quote(string(password)), `"`)
	//time.Sleep(100 * time.Millisecond)
	return nil, nil
}

func (s *SSHSession) HandleConn(conn net.Conn, serverConfig *ssh.ServerConfig) {
	defer conn.Close()
	s.Service = "ssh"
	//(conn.RemoteAddr())
	sshConn, chans, Request, err := ssh.NewServerConn(conn, serverConfig)
	if err != nil {
		s.Error = true
		//log.Println("Failed to handshake:", err)
		return
	}
	s.ClientVersion = strings.Trim(strconv.Quote(string(sshConn.ClientVersion())), `"`)

	go func() {
		for req := range Request {

			s.Request += fmt.Sprintf("%s %s\n", req.Type, strings.Trim(strconv.Quote(string(req.Payload)), `"`))
		}
	}()

	//log.Printf("New SSH connection from %s\n", sshConn.RemoteAddr())

	//go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		go s.handleChannel(newChannel)
	}
}

type ConnLogger struct {
	io.ReadWriter
	in  *bytes.Buffer
	out *bytes.Buffer
}

// NewConnLogger 创建一个 ConnLogger 实例

// 重写 net.Conn 的 Read 方法，记录输入流量
func (cl *ConnLogger) Read(b []byte) (int, error) {
	n, err := cl.ReadWriter.Read(b)
	if n > 0 {
		cl.in.Write(b[:n])
	}
	return n, err
}

// 重写 net.Conn 的 Write 方法，记录输出流量
func (cl *ConnLogger) Write(b []byte) (int, error) {
	n, err := cl.ReadWriter.Write(b)
	if n > 0 {
		_, err := cl.out.Write(b[:n])
		if err != nil {

			fmt.Println(err)
		}

	}
	return n, err
}
func NewConnLogger(conn io.ReadWriter, in *bytes.Buffer, out *bytes.Buffer) *ConnLogger {
	return &ConnLogger{ReadWriter: conn, in: in, out: out}
}
func (s *SSHSession) handleChannel(newChannel ssh.NewChannel) {
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}
	//NewTerminal(newChannel)
	channel, Request, err := newChannel.Accept()
	if err != nil {
		return
	}
	var ch chan struct{}
	ch = make(chan struct{})
	var inBuffer bytes.Buffer
	var outBuffer bytes.Buffer
	channelLogger := NewConnLogger(channel, &inBuffer, &outBuffer)
	go func() {
		for req := range Request {

			if req.Type == "shell" {
				ch <- struct{}{}
				s.IsInteract = true
			} else if req.Type == "exec" {
				var payload = struct{ Value string }{}
				err := ssh.Unmarshal(req.Payload, &payload)
				if err != nil {
					s.Error = true
					s.Shell += strings.Trim(strconv.Quote(string(req.Payload)), `"`)
				} else {
					s.Shell += payload.Value
					if strings.Contains(payload.Value, "scp -t") {
						//欺骗SCP 客户端
						channel.Write([]byte("\x00"))
						reader := bufio.NewReader(channelLogger)
						_, err := reader.ReadString('\n')
						if err != nil {
							fmt.Println("Error reading:", err)
							break
						}
						channel.Write([]byte("\x00"))
						io.ReadAll(channelLogger)
						continue
					}
					if strings.Contains(payload.Value, "echo") {
						channel.Write([]byte(payload.Value + "\n"))
						continue
					}

					if strings.Contains(payload.Value, "uname -s -v -n -r -m") {
						channel.Write([]byte("Linux ubuntu 3.13.0-24-generic #47-Ubuntu SMP Fri May 2 23:30:00 UTC 2014 x86_64\n"))
					}

					if strings.Contains(payload.Value, "uname -a") {
						channel.Write([]byte("Linux ubuntu 3.13.0-24-generic #47-Ubuntu SMP Fri May 2 23:30:00 UTC 2014 x86_64 x86_64 x86_64 GNU/Linux\n"))
					}
					if strings.Contains(payload.Value, "whoami") {
						channel.Write([]byte("root\r"))
					}
					if strings.Contains(payload.Value, "id") {
						channel.Write([]byte("uid=0(root) gid=0(root) groups=0(root) context=unconfined_u:unconfined_r:unconfined_t:s0-s0:c0.c1023\n"))
					}

				}
			}
			s.Request += fmt.Sprintf("%s %s\n", req.Type, strings.Trim(strconv.Quote(string(req.Payload)), `"`))
		}
		close(ch)
	}()
	<-ch
	defer channel.Close()

	defer func() { s.Data = strings.Trim(strconv.Quote(string(inBuffer.Bytes())), `"`) }()

	term := terminal.NewTerminal(channelLogger, "[root@ubuntu ~]# ")

	for {
		// get user input
		line, err := term.ReadLine()
		if err != nil {
			break
		}
		s.Shell += fmt.Sprintf("%s\n", strings.Trim(strconv.Quote(line), `"`))
		if len(line) > 0 {
			switch strings.Fields(line)[0] {
			case "exit":
				return
			case "env":
				term.Write([]byte("SHELL=/bin/bash\nUSER=root\nPATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\nPWD=/root\nLANG=en_US.UTF-8\nSHLVL=1\nHOME=/root\nLOGNAME=root"))
			case "ls":
				term.Write([]byte("Desktop  Documents  Downloads  Music  Pictures  Public  Templates  Videos\n"))
			case "uname":
				if len(strings.Fields(line)) > 1 {
					switch strings.Fields(line)[1] {
					case "-a":
						term.Write([]byte("Linux ubuntu 3.13.0-24-generic #47-Ubuntu SMP Fri May 2 23:30:00 UTC 2014 x86_64 x86_64 x86_64 GNU/Linux\n"))
					case "-i":
						term.Write([]byte("x86_64\n"))
					case "-p":
						term.Write([]byte("x86_64\n"))

					case "-o":
						term.Write([]byte("GNU/Linux\n"))
					default:
						term.Write([]byte("x86_64\n"))

					}

				} else {
					term.Write([]byte("Linux\n"))
				}
			case "whoami":
				term.Write([]byte("root\n"))
			case "hostname":
				term.Write([]byte("ubuntu\n"))

			case "id":
				term.Write([]byte("uid=0(root) gid=0(root) groups=0(root) context=unconfined_u:unconfined_r:unconfined_t:s0-s0:c0.c1023\n"))
			case "echo":
				if len(strings.Fields(line)) > 1 {
					term.Write([]byte(strings.Join(strings.Fields(line)[1:], " ") + "\n"))
				}
			case "pwd":
				term.Write([]byte("/root\n"))
			case "ps":
				term.Write([]byte("PID TTY          TIME CMD\n1 ?        00:00:00 init\n2 ?        00:00:00 kthreadd\n3 ?        00:00:00 ksoftirqd/0\n5 ?        00:00:00 kworker/0:0H\n7 ?        00:00:00 rcu_sched\n8 ?        00:00:00 rcu_bh\n9 ?        00:00:00 migration/0\n10 ?        00:00:00 watchdog/0\n11 ?        00:00:00 cpuhp/0\n12 ?        00:00:00 kdevtmpfs\n13 ?        00:00:00 netns\n14 ?        00:00:00 khungtaskd\n15 ?        00:00:00 oom_reaper\n16 ?        00:00:00 writeback\n17 ?        00:00:00 kcompactd0\n18 ?        00:00:00 ksmd\n19 ?        00:00:00 khugepaged\n20 ?        00:00:00 crypto\n21 ?        00:00:00 kintegrityd\n22 ?        00:00:00 kblockd\n23 ?        00:00:00 ata_sff\n24 ?        00:00:00 md\n25 ?        00:00:00 edac-poller\n26 ?        00:00:00 devfreq_wq\n27 ?        00:00:00 watchdogd\n28 ?        00:00:00 kswapd0\n29 ?        00:00:00 bash\n"))
			case "ifconfig":
				term.Write([]byte("eth0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500\n        inet 10.45.14.41  netmask 255.255.255.0  broadcast 10.45.14.255\n        ether 00:01:11:45:cd:88  txqueuelen 1000  (Ethernet)\n        RX packets 69959146  bytes 4174629619 (4.7 GB)\n        RX errors 0  dropped 0  overruns 0  frame 0\n        TX packets 92871777  bytes 7929333630 (7.2 GB)\n        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0\n\nlo: flags=73<UP,LOOPBACK,RUNNING>  mtu 65536\n        inet 127.0.0.1  netmask 255.0.0.0\n        inet6 ::1  prefixlen 128  scopeid 0x10<host>\n        loop  txqueuelen 1000  (Local Loopback)\n        RX packets 324635  bytes 24156673 (24.1 MB)\n        RX errors 0  dropped 488  overruns 0  frame 0\n        TX packets 324635  bytes 24156673 (24.1 MB)\n        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0\n\n"))
			default:

			}
		}
	}

}
