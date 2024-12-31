package rules

import (
	"crypto/sha1"
	"encoding/hex"
	"strconv"
	"strings"
)

func boolToUint8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

func calculateSHA1(input string) string {

	hasher := sha1.New()
	hasher.Write([]byte(input))
	hashSum := hasher.Sum(nil)

	hashString := hex.EncodeToString(hashSum)
	return hashString
}
func unquote(data string) []byte {
	s, err := strconv.Unquote(`"` + data + `"`)
	if err != nil {
		return []byte(data)
	}
	return []byte(s)
}

func indexOfH(data []byte, hexString string) int {
	hexBytes, err := ParseHex(hexString)
	if err != nil {
		return -1
	}

	for i := 0; i <= len(data)-len(hexBytes); i++ {
		if compareBytes(data[i:i+len(hexBytes)], hexBytes) {
			return i
		}
	}

	return -1
}

func lastIndexOfH(data []byte, hexString string) int {
	hexBytes, err := ParseHex(hexString)
	if err != nil {
		return -1
	}

	for i := len(data) - len(hexBytes); i >= 0; i-- {
		if compareBytes(data[i:i+len(hexBytes)], hexBytes) {
			return i
		}
	}

	return -1
}

func hasPrefixH(data []byte, hexString string) bool {
	hexBytes, err := ParseHex(hexString)
	if err != nil || len(data) < len(hexBytes) {
		return false
	}

	return compareBytes(data[:len(hexBytes)], hexBytes)
}

func hasSuffixH(data []byte, hexString string) bool {
	hexBytes, err := ParseHex(hexString)
	if err != nil || len(data) < len(hexBytes) {
		return false
	}

	return compareBytes(data[len(data)-len(hexBytes):], hexBytes)
}
func compareBytesH(a []byte, hexString string) bool {
	b, err := ParseHex(hexString)
	if err != nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
func compareBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func ParseHex(hexString string) ([]byte, error) {
	hexString = strings.ReplaceAll(hexString, " ", "")
	hexString = strings.ReplaceAll(hexString, "\\x", "")
	// Decode the hexadecimal string
	decoded, err := hex.DecodeString(hexString)
	if err != nil {
		return nil, err
	}

	return decoded, nil
}
