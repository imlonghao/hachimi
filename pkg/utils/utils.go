package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"strings"
	"time"
)

type NewConn struct {
	net.Conn
	Reader  io.Reader
	Writer  io.Writer
	Counter int64
	limit   int64
}

func NewLoggedConn(conn net.Conn, reader io.Reader, logWriter io.Writer, limit int64) *NewConn {
	if logWriter == nil {
		logWriter = io.Discard // 默认丢弃所有日志
	}
	return &NewConn{
		Conn:   conn,
		Reader: io.TeeReader(reader, logWriter), // 记录读取流量
		Writer: conn,
		limit:  limit,
	}
}

func (c *NewConn) Read(p []byte) (int, error) {
	// 全局读取限制
	if c.limit != 0 && c.Counter >= c.limit {
		//EOF
		return 0, io.EOF
	}
	n, err := c.Reader.Read(p)
	c.Counter += int64(n)
	return n, err
}

func (c *NewConn) Close() error {
	// 所有的连接关闭在最外层控制 防止被其他组件意外关闭
	return nil
}

// EscapeBytes escapes a byte slice to a string, using C-style escape sequences.
func EscapeBytes(data []byte) string {
	lowerhex := "0123456789abcdef"
	var builder strings.Builder
	builder.Grow(len(data) * 2) // 预分配空间

	for _, b := range data {
		if b == '\\' { // always backslashed
			builder.WriteByte('\\')
			builder.WriteByte(b)
			continue
		}

		switch b {
		case '\a':
			builder.WriteString(`\a`)
		case '\b':
			builder.WriteString(`\b`)
		case '\f':
			builder.WriteString(`\f`)
		case '\n':
			builder.WriteString(`\n`)
		case '\r':
			builder.WriteString(`\r`)
		case '\t':
			builder.WriteString(`\t`)
		case '\v':
			builder.WriteString(`\v`)
		default:
			switch {
			case b < 0x20 || b == 0x7f || b >= 0x80:
				builder.WriteString(`\x`)
				builder.WriteByte(lowerhex[b>>4])
				builder.WriteByte(lowerhex[b&0xF])
			default:
				builder.WriteByte(b)
			}
		}
	}
	return builder.String()
}

// UnescapeBytes unescapes a string to a byte slice, using C-style escape sequences.
func UnescapeBytes(data string) ([]byte, error) {
	var result []byte
	for i := 0; i < len(data); i++ {
		if data[i] == '\\' {
			i++ // Skip the backslash
			if i >= len(data) {
				return nil, fmt.Errorf("invalid escape sequence at end of string")
			}
			switch data[i] {
			case 'a':
				result = append(result, '\a')
			case 'b':
				result = append(result, '\b')
			case 'f':
				result = append(result, '\f')
			case 'n':
				result = append(result, '\n')
			case 'r':
				result = append(result, '\r')
			case 't':
				result = append(result, '\t')
			case 'v':
				result = append(result, '\v')
			case 'x':
				if i+2 >= len(data) {
					return nil, fmt.Errorf("invalid \\x escape sequence")
				}
				high := decodeHex(data[i+1])
				low := decodeHex(data[i+2])
				if high < 0 || low < 0 {
					return nil, fmt.Errorf("invalid hex digit in \\x escape sequence")
				}
				result = append(result, byte(high<<4|low))
				i += 2
			case 'u':
				if i+4 >= len(data) {
					return nil, fmt.Errorf("invalid \\u escape sequence")
				}
				var r rune
				for j := 0; j < 4; j++ {
					v := decodeHex(data[i+1+j])
					if v < 0 {
						return nil, fmt.Errorf("invalid hex digit in \\u escape sequence")
					}
					r = r<<4 | rune(v)
				}
				result = append(result, string(r)...)
				i += 4
			case '\\':
				result = append(result, '\\')
			default:
				return nil, fmt.Errorf("unknown escape sequence: \\%c", data[i])
			}
		} else {
			result = append(result, data[i])
		}
	}
	return result, nil
}

// decodeHex decodes a hexadecimal digit.
func decodeHex(b byte) int {
	switch {
	case '0' <= b && b <= '9':
		return int(b - '0')
	case 'a' <= b && b <= 'f':
		return int(b - 'a' + 10)
	case 'A' <= b && b <= 'F':
		return int(b - 'A' + 10)
	default:
		return -1
	}
}
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ReadAll 读取所有数据 直到达到限制 返回是否超出限制
func ReadAll(conn net.Conn, limit int64) bool {
	var buffer = make([]byte, 1024)
	var total int64
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return false
		}
		total += int64(n)
		if total >= limit {
			return true
		}
	}
}

// ToMap 将结构体转为 map 不支持嵌套结构体
func ToMap(obj interface{}) (map[string]interface{}, error) {
	// 如果 obj 是 nil 或者不是结构体，直接返回错误
	if obj == nil {
		return nil, errors.New("input object is nil")
	}
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem() // 解引用指针
	}
	if val.Kind() != reflect.Struct {
		return nil, errors.New("input is not a struct")
	}
	// 遍历结构体字段，将其转为 map
	result := make(map[string]interface{})
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			// 跳过未导出的字段
			continue
		}
		result[field.Name] = val.Field(i).Interface()
	}
	return result, nil
}

// StringToTime time.Time字符串转为 time.Time
func StringToTime(s string) (time.Time, error) {
	dateTime, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		dateTime, err = time.Parse(time.RFC3339, s)
	}
	return dateTime, err

}

// 	map[string]interface {} to  map[string]string

func MapInterfaceToString(m map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		if str, ok := v.(string); ok {
			result[k] = str
		}
	}
	return result
}
func SHA1(input string) string {

	hasher := sha1.New()
	hasher.Write([]byte(input))
	hashSum := hasher.Sum(nil)

	hashString := hex.EncodeToString(hashSum)
	return hashString
}
