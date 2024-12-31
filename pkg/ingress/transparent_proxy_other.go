//go:build !linux
// +build !linux

package ingress

import (
	"errors"
	"net"
	"os"
)

func TransparentProxy(file *os.File) error {
	return errors.New("TransparentProxy not supported on this platform")
}
func getOrigDst(oob []byte, oobn int) (*net.UDPAddr, error) {
	return nil, errors.New("getOrigDst not supported on this platform")
}
