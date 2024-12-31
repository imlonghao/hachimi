//go:build linux
// +build linux

package ingress

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"os"
	"syscall"
	"unsafe"
)

func TransparentProxy(file *os.File) error {
	fd := int(file.Fd())
	if err := syscall.SetsockoptInt(fd, syscall.SOL_IP, syscall.IP_TRANSPARENT, 1); err != nil {
		return err
	}
	if err := syscall.SetsockoptInt(fd, syscall.SOL_IP, syscall.IP_RECVORIGDSTADDR, 1); err != nil {
		return err
	}
	return nil
}

// 一些在 Go 中尚未定义/导出的常量，手动补齐
const (
	// IP_RECVORIGDSTADDR IPv4 原始目的地址常量
	IP_RECVORIGDSTADDR = 0x14

	// IPV6_RECVORIGDSTADDR IPv6 原始目的地址常量
	IPV6_RECVORIGDSTADDR = 0x4a
)

// getOrigDst 提取 UDP 原始目的地址
func getOrigDst(oob []byte, oobn int) (*net.UDPAddr, error) {
	msgs, err := syscall.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		return nil, err
	}

	var origDst *net.UDPAddr

	for _, msg := range msgs {
		switch {
		// ===================== IPv4 =====================
		case msg.Header.Level == syscall.SOL_IP && msg.Header.Type == IP_RECVORIGDSTADDR:
			origDstRaw := &syscall.RawSockaddrInet4{}
			if err := binary.Read(bytes.NewReader(msg.Data), binary.LittleEndian, origDstRaw); err != nil {
				return nil, err
			}
			if origDstRaw.Family != syscall.AF_INET {
				return nil, errors.New("unsupported family for IPv4")
			}

			pp := (*syscall.RawSockaddrInet4)(unsafe.Pointer(origDstRaw))
			p := (*[2]byte)(unsafe.Pointer(&pp.Port))

			origDst = &net.UDPAddr{
				IP:   net.IPv4(pp.Addr[0], pp.Addr[1], pp.Addr[2], pp.Addr[3]),
				Port: int(p[0])<<8 + int(p[1]),
			}

		// ===================== IPv6 (TProxy) =====================
		case msg.Header.Level == syscall.SOL_IPV6 && msg.Header.Type == IPV6_RECVORIGDSTADDR:
			origDstRaw := &syscall.RawSockaddrInet6{}
			if err := binary.Read(bytes.NewReader(msg.Data), binary.LittleEndian, origDstRaw); err != nil {
				return nil, err
			}
			if origDstRaw.Family != syscall.AF_INET6 {
				return nil, errors.New("unsupported family for IPv6")
			}

			pp := (*syscall.RawSockaddrInet6)(unsafe.Pointer(origDstRaw))
			p := (*[2]byte)(unsafe.Pointer(&pp.Port))

			origDst = &net.UDPAddr{
				IP:   net.IP(pp.Addr[:]),
				Port: int(p[0])<<8 + int(p[1]),
				Zone: zoneFromIndex(pp.Scope_id),
			}
		}
	}

	if origDst == nil {
		return nil, errors.New("original destination not found in control messages")
	}

	return origDst, nil
}

// zoneFromIndex 将作用域 ID 转换为网卡名称，比如 "eth0"。
func zoneFromIndex(index uint32) string {
	ifi, err := net.InterfaceByIndex(int(index))
	if err != nil {
		return ""
	}
	return ifi.Name
}
