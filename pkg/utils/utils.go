package utils

import (
	"fmt"
	"io"
	"net"
	"strings"
)

type NewConn struct {
	net.Conn
	Reader io.Reader
	Writer io.Writer
}

func NewLoggedConn(conn net.Conn, reader io.Reader, logWriter io.Writer) *NewConn {
	if logWriter == nil {
		logWriter = io.Discard // 默认丢弃所有日志
	}
	return &NewConn{
		Conn:   conn,
		Reader: io.TeeReader(reader, logWriter), // 记录读取流量
		Writer: conn,
	}
}

func (c *NewConn) Read(p []byte) (int, error) {
	return c.Reader.Read(p)
}

func (c *NewConn) Close() error {
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
