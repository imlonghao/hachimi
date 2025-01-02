package redis

import (
	"bufio"
	"fmt"
	"github.com/google/uuid"
	"hachimi/pkg/config"
	"hachimi/pkg/types"
	"hachimi/pkg/utils"
	"io"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

func HandleRedis(conn net.Conn, session *types.Session) bool {
	c := newConn(conn, config.GetConfig().TimeOut)
	list := map[string]bool{"acl": true,
		"append":               true,
		"asking":               true,
		"auth":                 true,
		"bf.add":               true,
		"bf.card":              true,
		"bf.exists":            true,
		"bf.info":              true,
		"bf.insert":            true,
		"bf.loadchunk":         true,
		"bf.madd":              true,
		"bf.mexists":           true,
		"bf.reserve":           true,
		"bf.scandump":          true,
		"bgrewriteaof":         true,
		"bgsave":               true,
		"bitcount":             true,
		"bitfield":             true,
		"bitfield_ro":          true,
		"bitop":                true,
		"bitpos":               true,
		"blmove":               true,
		"blmpop":               true,
		"blpop":                true,
		"brpop":                true,
		"brpoplpush":           true,
		"bzmpop":               true,
		"bzpopmax":             true,
		"bzpopmin":             true,
		"cf.add":               true,
		"cf.addnx":             true,
		"cf.count":             true,
		"cf.del":               true,
		"cf.exists":            true,
		"cf.info":              true,
		"cf.insert":            true,
		"cf.insertnx":          true,
		"cf.loadchunk":         true,
		"cf.mexists":           true,
		"cf.reserve":           true,
		"cf.scandump":          true,
		"client":               true,
		"cluster":              true,
		"cms.incrby":           true,
		"cms.info":             true,
		"cms.initbydim":        true,
		"cms.initbyprob":       true,
		"cms.merge":            true,
		"cms.query":            true,
		"command":              true,
		"config":               true,
		"copy":                 true,
		"dbsize":               true,
		"decr":                 true,
		"decrby":               true,
		"del":                  true,
		"discard":              true,
		"dump":                 true,
		"echo":                 true,
		"eval":                 true,
		"eval_ro":              true,
		"evalsha":              true,
		"evalsha_ro":           true,
		"exec":                 true,
		"exists":               true,
		"expire":               true,
		"expireat":             true,
		"expiretime":           true,
		"failover":             true,
		"fcall":                true,
		"fcall_ro":             true,
		"flushall":             true,
		"flushdb":              true,
		"ft._list":             true,
		"ft.aggregate":         true,
		"ft.aliasadd":          true,
		"ft.aliasdel":          true,
		"ft.aliasupdate":       true,
		"ft.alter":             true,
		"ft.config":            true,
		"ft.create":            true,
		"ft.cursor":            true,
		"ft.dictadd":           true,
		"ft.dictdel":           true,
		"ft.dictdump":          true,
		"ft.dropindex":         true,
		"ft.explain":           true,
		"ft.explaincli":        true,
		"ft.info":              true,
		"ft.profile":           true,
		"ft.search":            true,
		"ft.spellcheck":        true,
		"ft.sugadd":            true,
		"ft.sugdel":            true,
		"ft.sugget":            true,
		"ft.suglen":            true,
		"ft.syndump":           true,
		"ft.synupdate":         true,
		"ft.tagvals":           true,
		"function":             true,
		"geoadd":               true,
		"geodist":              true,
		"geohash":              true,
		"geopos":               true,
		"georadius":            true,
		"georadius_ro":         true,
		"georadiusbymember":    true,
		"georadiusbymember_ro": true,
		"geosearch":            true,
		"geosearchstore":       true,
		"get":                  true,
		"getbit":               true,
		"getdel":               true,
		"getex":                true,
		"getrange":             true,
		"getset":               true,
		"hdel":                 true,
		"hello":                true,
		"hexists":              true,
		"hget":                 true,
		"hgetall":              true,
		"hincrby":              true,
		"hincrbyfloat":         true,
		"hkeys":                true,
		"hlen":                 true,
		"hmget":                true,
		"hmset":                true,
		"hrandfield":           true,
		"hscan":                true,
		"hset":                 true,
		"hsetnx":               true,
		"hstrlen":              true,
		"hvals":                true,
		"incr":                 true,
		"incrby":               true,
		"incrbyfloat":          true,
		"info":                 true,
		"json.arrappend":       true,
		"json.arrindex":        true,
		"json.arrinsert":       true,
		"json.arrlen":          true,
		"json.arrpop":          true,
		"json.arrtrim":         true,
		"json.clear":           true,
		"json.debug":           true,
		"json.del":             true,
		"json.forget":          true,
		"json.get":             true,
		"json.merge":           true,
		"json.mget":            true,
		"json.mset":            true,
		"json.numincrby":       true,
		"json.nummultby":       true,
		"json.objkeys":         true,
		"json.objlen":          true,
		"json.resp":            true,
		"json.set":             true,
		"json.strappend":       true,
		"json.strlen":          true,
		"json.toggle":          true,
		"json.type":            true,
		"keys":                 true,
		"lastsave":             true,
		"latency":              true,
		"lcs":                  true,
		"lindex":               true,
		"linsert":              true,
		"llen":                 true,
		"lmove":                true,
		"lmpop":                true,
		"lolwut":               true,
		"lpop":                 true,
		"lpos":                 true,
		"lpush":                true,
		"lpushx":               true,
		"lrange":               true,
		"lrem":                 true,
		"lset":                 true,
		"ltrim":                true,
		"memory":               true,
		"mget":                 true,
		"migrate":              true,
		"module":               true,
		"monitor":              true,
		"move":                 true,
		"mset":                 true,
		"msetnx":               true,
		"multi":                true,
		"object":               true,
		"persist":              true,
		"pexpire":              true,
		"pexpireat":            true,
		"pexpiretime":          true,
		"pfadd":                true,
		"pfcount":              true,
		"pfdebug":              true,
		"pfmerge":              true,
		"pfselftest":           true,
		"ping":                 true,
		"psetex":               true,
		"psubscribe":           true,
		"psync":                true,
		"pttl":                 true,
		"publish":              true,
		"pubsub":               true,
		"punsubscribe":         true,
		"quit":                 true,
		"randomkey":            true,
		"readonly":             true,
		"readwrite":            true,
		"rename":               true,
		"renamenx":             true,
		"replconf":             true,
		"replicaof":            true,
		"reset":                true,
		"restore":              true,
		"restore-asking":       true,
		"role":                 true,
		"rpop":                 true,
		"rpoplpush":            true,
		"rpush":                true,
		"rpushx":               true,
		"sadd":                 true,
		"save":                 true,
		"scan":                 true,
		"scard":                true,
		"script":               true,
		"sdiff":                true,
		"sdiffstore":           true,
		"select":               true,
		"set":                  true,
		"setbit":               true,
		"setex":                true,
		"setnx":                true,
		"setrange":             true,
		"shutdown":             true,
		"sinter":               true,
		"sintercard":           true,
		"sinterstore":          true,
		"sismember":            true,
		"slaveof":              true,
		"slowlog":              true,
		"smembers":             true,
		"smismember":           true,
		"smove":                true,
		"sort":                 true,
		"sort_ro":              true,
		"spop":                 true,
		"spublish":             true,
		"srandmember":          true,
		"srem":                 true,
		"sscan":                true,
		"ssubscribe":           true,
		"strlen":               true,
		"subscribe":            true,
		"substr":               true,
		"sunion":               true,
		"sunionstore":          true,
		"sunsubscribe":         true,
		"swapdb":               true,
		"sync":                 true,
		"tdigest.add":          true,
		"tdigest.byrank":       true,
		"tdigest.byrevrank":    true,
		"tdigest.cdf":          true,
		"tdigest.create":       true,
		"tdigest.info":         true,
		"tdigest.max":          true,
		"tdigest.merge":        true,
		"tdigest.min":          true,
		"tdigest.quantile":     true,
		"tdigest.rank":         true,
		"tdigest.reset":        true,
		"tdigest.revrank":      true,
		"tdigest.trimmed_mean": true,
		"tfcall":               true,
		"tfcallasync":          true,
		"tfunction":            true,
		"time":                 true,
		"topk.add":             true,
		"topk.count":           true,
		"topk.incrby":          true,
		"topk.info":            true,
		"topk.list":            true,
		"topk.query":           true,
		"topk.reserve":         true,
		"touch":                true,
		"ts.add":               true,
		"ts.alter":             true,
		"ts.create":            true,
		"ts.createrule":        true,
		"ts.decrby":            true,
		"ts.del":               true,
		"ts.deleterule":        true,
		"ts.get":               true,
		"ts.incrby":            true,
		"ts.info":              true,
		"ts.madd":              true,
		"ts.mget":              true,
		"ts.mrange":            true,
		"ts.mrevrange":         true,
		"ts.queryindex":        true,
		"ts.range":             true,
		"ts.revrange":          true,
		"ttl":                  true,
		"type":                 true,
		"unlink":               true,
		"unsubscribe":          true,
		"unwatch":              true,
		"wait":                 true,
		"waitaof":              true,
		"watch":                true,
		"xack":                 true,
		"xadd":                 true,
		"xautoclaim":           true,
		"xclaim":               true,
		"xdel":                 true,
		"xgroup":               true,
		"xinfo":                true,
		"xlen":                 true,
		"xpending":             true,
		"xrange":               true,
		"xread":                true,
		"xreadgroup":           true,
		"xrevrange":            true,
		"xsetid":               true,
		"xtrim":                true,
		"zadd":                 true,
		"zcard":                true,
		"zcount":               true,
		"zdiff":                true,
		"zdiffstore":           true,
		"zincrby":              true,
		"zinter":               true,
		"zintercard":           true,
		"zinterstore":          true,
		"zlexcount":            true,
		"zmpop":                true,
		"zmscore":              true,
		"zpopmax":              true,
		"zpopmin":              true,
		"zrandmember":          true,
		"zrange":               true,
		"zrangebylex":          true,
		"zrangebyscore":        true,
		"zrangestore":          true,
		"zrank":                true,
		"zrem":                 true,
		"zremrangebylex":       true,
		"zremrangebyrank":      true,
		"zremrangebyscore":     true,
		"zrevrange":            true,
		"zrevrangebylex":       true,
		"zrevrangebyscore":     true,
		"zrevrank":             true,
		"zscan":                true,
		"zscore":               true,
		"zunion":               true}
	sess := &types.RedisSession{}
	sess.ID = uuid.New().String()
	sess.SessionID = session.ID
	sess.Service = "redis"
	sess.StartTime = time.Now()
	defer func() {
		sess.EndTime = time.Now()
		sess.Duration = int(sess.EndTime.Sub(sess.StartTime).Milliseconds())
		config.Logger.Log(sess)
	}()
	for {
		request, err := parseRequest(c)
		if err != nil {
			sess.Error = true
			if err == io.EOF {
				return true
			} else {
				if len(sess.Data) == 0 {
					return false //降级到非redis
				}
			}
			return true

		} else if strings.ToLower(request.Name) == "auth" {
			if len(request.Args) == 1 {
				sess.PassWord = string(request.Args[0])
			} else if len(request.Args) == 2 {
				sess.User = string(request.Args[0])
				sess.PassWord = string(request.Args[1])
			}

			conn.Write([]byte("+OK\r\n"))
		} else if strings.ToLower(request.Name) == "info" {
			if len(request.Args) > 0 {
				if strings.ToLower(string(request.Args[0])) == "server" {
					conn.Write([]byte("$439\r\n# Server\r\nredis_version:4.0.8\r\nredis_git_sha1:00000000\r\nredis_git_dirty:0\r\nredis_build_id:c2238b38b1edb0e2\r\nredis_mode:standalone\r\nos:Linux 3.10.0-1024.1.2.el7.x86_64 x86_64\r\narch_bits:64\r\nmultiplexing_api:epoll\r\ngcc_version:4.8.5\r\nprocess_id:3772\r\nrun_id:0e61abd297771de3fe812a3c21027732ac9f41fe\r\ntcp_port:6379\r\nuptime_in_seconds:25926381\r\nuptime_in_days:300\r\nhz:10\r\nlru_clock:13732392\r\nconfig_file:/usr/local/redis-local/etc/redis.conf\r\n\r\n"))
					continue
				}
			}

			conn.Write([]byte("$2012\r\n# Server\r\nredis_version:4.0.8\r\nredis_git_sha1:00000000\r\nredis_git_dirty:0\r\nredis_build_id:ca4ed916473088db\r\nredis_mode:standalone\r\nos:Linux 3.10.0-1024.1.2.el7.x86_64 x86_64\r\narch_bits:64\r\nmultiplexing_api:epoll\r\ngcc_version:4.8.5\r\nprocess_id:3772\r\nrun_id:0e61abd297771de3fe812a3c21027732ac9f41fe\r\ntcp_port:6379\r\nuptime_in_seconds:25926381\r\nuptime_in_days:300\r\nhz:10\r\nlru_clock:13732392\r\nconfig_file:/usr/local/redis-local/etc/redis.conf\r\n\r\n# Clients\r\nconnected_clients:208\r\nclient_longest_output_list:0\r\nclient_biggest_input_buf:517\r\nblocked_clients:0\r\n\r\n# Memory\r\nused_memory:5151720\r\nused_memory_human:4.91M\r\nused_memory_rss:6885376\r\nused_memory_peak:5214456\r\nused_memory_peak_human:4.97M\r\nused_memory_lua:61440\r\nmem_fragmentation_ratio:1.34\r\nmem_allocator:jemalloc-3.6.0\r\n\r\n# Persistence\r\nloading:0\r\nrdb_changes_since_last_save:0\r\nrdb_bgsave_in_progress:0\r\nrdb_last_save_time:1704104232\r\nrdb_last_bgsave_status:ok\r\nrdb_last_bgsave_time_sec:0\r\nrdb_current_bgsave_time_sec:-1\r\naof_enabled:0\r\naof_rewrite_in_progress:0\r\naof_rewrite_scheduled:0\r\naof_last_rewrite_time_sec:-1\r\naof_current_rewrite_time_sec:-1\r\naof_last_bgrewrite_status:ok\r\naof_last_write_status:ok\r\n\r\n# Stats\r\ntotal_connections_received:509000\r\ntotal_commands_processed:616946\r\ninstantaneous_ops_per_sec:0\r\ntotal_net_input_bytes:20893857\r\ntotal_net_output_bytes:39299490\r\ninstantaneous_input_kbps:0.00\r\ninstantaneous_output_kbps:0.00\r\nrejected_connections:0\r\nsync_full:0\r\nsync_partial_ok:0\r\nsync_partial_err:0\r\nexpired_keys:45\r\nevicted_keys:0\r\nkeyspace_hits:49026\r\nkeyspace_misses:308\r\npubsub_channels:0\r\npubsub_patterns:0\r\nlatest_fork_usec:300\r\nmigrate_cached_sockets:0\r\n\r\n# Replication\r\nrole:master\r\nconnected_slaves:0\r\nmaster_repl_offset:0\r\nrepl_backlog_active:0\r\nrepl_backlog_size:1048576\r\nrepl_backlog_first_byte_offset:0\r\nrepl_backlog_histlen:0\r\n\r\n# CPU\r\nused_cpu_sys:15257.43\r\nused_cpu_user:10518.80\r\nused_cpu_sys_children:279.62\r\nused_cpu_user_children:31.15\r\n\r\n# Cluster\r\ncluster_enabled:0\r\n\r\n# Keyspace\r\ndb0:keys=1,expires=0,avg_ttl=0\r\n\r\n"))

		} else if strings.ToLower(request.Name) == "ping" {
			if len(request.Args) > 0 {
				conn.Write([]byte("*" + strconv.Itoa(len(request.Args[0])) + "\r\n"))
				conn.Write(request.Args[0])
				conn.Write([]byte("\r\n"))
			} else {
				conn.Write([]byte("+PONG\r\n"))
			}
		} else if list[strings.ToLower(request.Name)] {
			conn.Write([]byte("+OK\r\n"))
		} else {
			conn.Write([]byte("-ERR unknown command '" + request.Name + "'\r\n")) //返回指纹特征 便于搜索引擎记录
			//if len(sess.Data) == 0 {
			//	return false //降级到非redis
			//}
		}
		var arg string
		arg = request.Name + " "
		for _, v := range request.Args {
			arg += utils.EscapeBytes(v) + " "
		}
		sess.Data += arg + "\n"

		//if _, err = reply.WriteTo(c.w); err != nil {
		//	return err
		//}
		/*
			if c.timeout > 0 {
				deadline := time.Now().Add(c.timeout)
				if err := c.nc.SetWriteDeadline(deadline); err != nil {
					return nil
				}
			}
		*/

	}
	return true
}

type Request struct {
	Name       string
	Args       [][]byte
	Host       string
	ClientChan chan struct{}
}

type Conn struct {
	r *bufio.Reader
	w *bufio.Writer

	wLock sync.Mutex

	db uint32
	nc net.Conn

	// summary for this connection
	summ    string
	timeout time.Duration

	authenticated bool

	// whether sync from master or not
	isSyncing bool
}

func newConn(nc net.Conn, timeout int) *Conn {
	c := &Conn{
		nc: nc,
	}

	c.r = bufio.NewReader(nc)
	c.w = bufio.NewWriter(nc)
	//c.summ = fmt.Sprintf("local%s-remote%s", nc.LocalAddr(), nc.RemoteAddr())
	c.timeout = time.Duration(timeout) * time.Second
	c.authenticated = false
	c.isSyncing = false
	//log.Info("connection established:", c.summ)

	return c
}

func (c *Conn) Close() {
	c.nc.Close()
	c = nil
}

func parseRequest(c *Conn) (*Request, error) {
	r := c.r
	// first line of redis request should be:
	// *<number of arguments>CRLF
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	// note that this line also protects us from negative integers
	var argsCount int

	// Multiline request:
	if line[0] == '*' {
		if _, err := fmt.Sscanf(line, "*%d\r", &argsCount); err != nil {
			return nil, malformed("*<numberOfArguments>", line)
		}
		// All next lines are pairs of:
		//$<number of bytes of argument 1> CR LF
		//<argument data> CR LF
		// first argument is a command name, so just convert
		firstArg, err := readArgument(r)
		if err != nil {
			return nil, err
		}

		args := make([][]byte, argsCount-1)
		for i := 0; i < argsCount-1; i += 1 {
			if args[i], err = readArgument(r); err != nil {
				return nil, err
			}
		}

		return &Request{
			Name: strings.ToLower(string(firstArg)),
			Args: args,
		}, nil
	}

	// Inline request:
	fields := strings.Split(strings.Trim(line, "\r\n"), " ")

	var args [][]byte
	if len(fields) > 1 {
		for _, arg := range fields[1:] {
			args = append(args, []byte(arg))
		}
	}
	return &Request{
		Name: strings.ToLower(string(fields[0])),
		Args: args,
	}, nil

}

func readArgument(r *bufio.Reader) ([]byte, error) {

	line, err := r.ReadString('\n')
	if err != nil {
		return nil, malformed("$<argumentLength>", line)
	}
	var argSize int
	if _, err := fmt.Sscanf(line, "$%d\r", &argSize); err != nil {
		return nil, malformed("$<argumentSize>", line)
	}

	// I think int is safe here as the max length of request
	// should be less then max int value?
	data, err := ioutil.ReadAll(io.LimitReader(r, int64(argSize)))
	if err != nil {
		return nil, err
	}

	if len(data) != argSize {
		return nil, malformedLength(argSize, len(data))
	}

	// Now check for trailing CR
	if b, err := r.ReadByte(); err != nil || b != '\r' {
		return nil, malformedMissingCRLF()
	}

	// And LF
	if b, err := r.ReadByte(); err != nil || b != '\n' {
		return nil, malformedMissingCRLF()
	}

	return data, nil
}

func malformed(expected string, got string) error {
	return fmt.Errorf("Mailformed request:'%s does not match %s\\r\\n'", got, expected)
}

func malformedLength(expected int, got int) error {
	return fmt.Errorf(
		"Mailformed request: argument length '%d does not match %d\\r\\n'",
		got, expected)
}

func malformedMissingCRLF() error {
	return fmt.Errorf("Mailformed request: line should end with \\r\\n")
}
