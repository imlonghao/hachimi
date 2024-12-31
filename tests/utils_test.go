package tests

import (
	"crypto/rand"
	"fmt"
	"github.com/savsgio/gotils/bytes"
	"hachimi/pkg/utils"
	"testing"
)

func TestEscapeBytes(t *testing.T) {
	//00-FF
	var data []byte
	for i := 0; i <= 255; i++ {
		data = append(data, byte(i))
	}
	fmt.Println(utils.EscapeBytes(data))
	unescapeBytes, err := utils.UnescapeBytes(utils.EscapeBytes(data))
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(unescapeBytes, data) {
		t.Error("UnescapeBytes error")
	}
	//随机10240
	data = make([]byte, 10240)
	rand.Read(data)
	unescapeBytes, err = utils.UnescapeBytes(utils.EscapeBytes(data))
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(unescapeBytes, data) {
		t.Error("UnescapeBytes error")
	}
}
