package tests

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"hachimi/pkg/config"
	"hachimi/pkg/ingress"
	"hachimi/pkg/logger"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"
)

type SimpleLineWriter struct {
	buffer bytes.Buffer
	mu     sync.Mutex
}

func (w *SimpleLineWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.Write(p)
}

func (w *SimpleLineWriter) ReadLines() []string {
	w.mu.Lock()
	defer w.mu.Unlock()

	var lines []string
	scanner := bufio.NewScanner(&w.buffer)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	// Clear the buffer after reading lines
	w.buffer.Reset()
	return lines
}

func NewSimpleLineWriter() *SimpleLineWriter {
	return &SimpleLineWriter{}
}

func TestListenerManager(t *testing.T) {
	// Test TCP Listeners
	writer := NewSimpleLineWriter()

	config.Logger = logger.NewJSONLLogger(writer, 100, "test")
	t.Run("Listeners", func(t *testing.T) {
		// Create a ListenerManager
		lm := ingress.NewListenerManager()
		// Add TCP listeners
		tcpListener := ingress.NewTCPListener("127.0.0.1", 54321)
		udpListener := ingress.NewUDPListener("127.0.0.1", 54321)

		lm.AddTCPListener(tcpListener)
		lm.AddUDPListener(udpListener)
		// Start all listeners
		lm.StartAll(context.Background())
		time.Sleep(1 * time.Second)
	})

	// Test HTTP and HTTPS Listeners
	t.Run("HTTP and HTTPS Log", func(t *testing.T) {

		specialParams := generateComprehensiveParams()
		binaryBody := generateRandBinary(1234)
		randomHeaders := generateRandomHeaders(5)
		randomPath := generateRandomString(10)
		// HTTP method
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
		//非标准请求
		methods = append(methods, generateRandomMethod())
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		} // Prepare the HTTP request
		// http/https
		scheme := []string{"http", "https"}
		for _, s := range scheme {
			for _, method := range methods {
				// Prepare the HTTP request

				testUrl := fmt.Sprintf("%s://127.0.0.1:54321/%s?%s", s, randomPath, specialParams)
				req, err := http.NewRequest(method, testUrl, bytes.NewReader(binaryBody))
				if err != nil {
					t.Fatalf("Failed to create HTTP request: %v", err)
				}
				// Add random headers to the request
				for k, v := range randomHeaders {
					req.Header.Add(k, v)
				}
				// Perform the HTTP request
				resp, err := client.Do(req)
				if err != nil {
					t.Fatalf("Method %s URL %s failed: %v", method, testUrl, err)
				}
				resp.Body.Close()
				// Log the response
				t.Logf("Response: %d %s", resp.StatusCode, resp.Status)
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Method %s URL %s failed with status code %d", method, testUrl, resp.StatusCode)
				}
			}
			config.Logger.Flush()
			logs := writer.ReadLines()
			if len(logs) != len(methods)*2 {
				t.Errorf("Expected %d log lines, got %d", len(methods)*2, len(logs))
			}
			//for _, logData := range logs {
			//TODO HTTP	日志完整性测试
			//}
			t.Logf("%s test passed", s)
		}
		t.Logf("HTTP and HTTPS test passed")
	})
}

// Helper function to generate all special URL characters and binary data as parameters
func generateComprehensiveParams() string {
	// Define all special URL characters
	specialChars := `!#$&'()*+,/:;=?@[]%`

	// Generate a key-value pair using all special characters and binary data
	key := url.QueryEscape("specialChars")

	value := url.QueryEscape(specialChars + string(generateFullBinary()))

	return fmt.Sprintf("%s=%s", key, value)
}

// Helper function to generate binary data containing all 0x00-0xFF bytes
func generateFullBinary() []byte {
	body := make([]byte, 256)
	for i := 0; i < 256; i++ {
		body[i] = byte(i)
	}
	return body
}

// Helper function to generate rand binary data
func generateRandBinary(len int) []byte {
	body := make([]byte, len)
	rand.Read(body)
	return body
}

func generateRandomHeaders(count int) map[string]string {
	headers := map[string]string{}
	for i := 0; i < count; i++ {
		key := generateRandomString(5)
		value := generateRandomString(10)
		headers[key] = value
	}
	return headers
}

func generateRandomString(length int) string {
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}
func generateRandomMethod() string {
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, 5)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}
