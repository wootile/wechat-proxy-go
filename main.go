package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// Version information
var Version = "dev"

// Allowed domain whitelist
var allowedDomains = []string{
	"api.weixin.qq.com",
	"api.wechat.com",
	"mp.weixin.qq.com",
	"qyapi.weixin.qq.com",
}

// Check if domain is allowed
func isDomainAllowed(host string) bool {
	// 记录原始 host
	originalHost := host
	
	// Remove port number
	if colonIndex := strings.LastIndex(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}

	// 添加详细日志
	log.Printf("检查域名权限 - 原始Host: %s, 清理后Host: %s", originalHost, host)

	for _, allowedDomain := range allowedDomains {
		// 仅完全匹配域名
		if host == allowedDomain {
			log.Printf("域名 %s 完全匹配白名单域名 %s", host, allowedDomain)
			return true
		}
	}
	
	log.Printf("域名 %s 不在白名单中，拒绝访问", host)
	log.Printf("当前白名单: %v", allowedDomains)
	return false
}

// Bidirectional data transfer
func copyData(dst, src net.Conn) {
	defer src.Close()
	defer dst.Close()
	io.Copy(dst, src)
}

// Handle HTTPS CONNECT requests
func handleConnect(w http.ResponseWriter, r *http.Request) {
	log.Printf("收到CONNECT请求 - Host: %s, Method: %s", r.Host, r.Method)
	
	if !isDomainAllowed(r.Host) {
		log.Printf("拒绝连接到未授权域名: %s", r.Host)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	log.Printf("允许连接到域名: %s", r.Host)

	// Connect to target server
	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		log.Printf("连接目标服务器失败: %v", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer destConn.Close()

	// Hijack HTTP connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		log.Printf("不支持连接劫持")
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		log.Printf("劫持连接失败: %v", err)
		return
	}
	defer clientConn.Close()

	// Send connection success response
	response := "HTTP/1.1 200 Connection Established\r\n\r\n"
	_, err = clientConn.Write([]byte(response))
	if err != nil {
		log.Printf("发送CONNECT响应失败: %v", err)
		return
	}

	log.Printf("成功建立CONNECT隧道到: %s", r.Host)

	// Bidirectional data transfer
	go func() {
		defer destConn.Close()
		defer clientConn.Close()
		_, err := io.Copy(destConn, clientConn)
		if err != nil {
			log.Printf("客户端->服务器数据传输错误: %v", err)
		}
	}()
	
	_, err = io.Copy(clientConn, destConn)
	if err != nil {
		log.Printf("服务器->客户端数据传输错误: %v", err)
	}
}

// Handle regular HTTP requests
func handleHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("收到HTTP请求 - Host: %s, Method: %s, URL: %s", r.Host, r.Method, r.URL.String())
	
	// Check if Host header exists
	if r.Host == "" {
		log.Printf("请求缺少Host头")
		http.Error(w, "Bad Request: missing Host header", http.StatusBadRequest)
		return
	}

	if !isDomainAllowed(r.Host) {
		log.Printf("拒绝HTTP请求到未授权域名: %s", r.Host)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	log.Printf("允许HTTP请求到域名: %s", r.Host)

	// Build target URL - fix URL construction logic
	var targetURL string
	if r.URL.IsAbs() {
		// If URL is already absolute, use it directly
		targetURL = r.URL.String()
		log.Printf("使用绝对URL: %s", targetURL)
	} else {
		// Build complete URL
		scheme := "http"
		// 对于代理服务器，通常我们需要根据端口判断协议
		// 如果Host包含端口443或者请求是通过TLS连接的，使用https
		if r.TLS != nil {
			scheme = "https"
		} else if strings.HasSuffix(r.Host, ":443") {
			scheme = "https"
		}
		
		// 确保RequestURI不为空
		requestURI := r.RequestURI
		if requestURI == "" {
			requestURI = r.URL.RequestURI()
		}
		
		targetURL = scheme + "://" + r.Host + requestURI
		log.Printf("构建相对URL: scheme=%s, host=%s, requestURI=%s -> %s", scheme, r.Host, requestURI, targetURL)
	}

	log.Printf("构建目标URL: %s", targetURL)

	// Create new request
	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		log.Printf("创建请求失败: %v", err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Copy request headers, but skip some proxy-related headers
	for key, values := range r.Header {
		// Skip headers that should not be forwarded
		if strings.ToLower(key) == "proxy-connection" || 
		   strings.ToLower(key) == "proxy-authenticate" || 
		   strings.ToLower(key) == "proxy-authorization" {
			continue
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Send request
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow redirects but limit the number
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
	
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("请求失败: %v", err)
		http.Error(w, "Request failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	log.Printf("收到响应 - 状态码: %d, 内容长度: %s", resp.StatusCode, resp.Header.Get("Content-Length"))

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy status code and response body
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("复制响应体失败: %v", err)
	}
}

// Unified handler function
func proxyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		handleConnect(w, r)
	} else {
		handleHTTP(w, r)
	}
}

func main() {
	// Handle command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("WeChat API Proxy Server v%s\n", Version)
			return
		case "--help", "-h":
			fmt.Printf("WeChat API Proxy Server v%s\n\n", Version)
			fmt.Printf("Usage: %s [options]\n\n", os.Args[0])
			fmt.Printf("Options:\n")
			fmt.Printf("  --version, -v    Show version information\n")
			fmt.Printf("  --help, -h       Show this help message\n\n")
			fmt.Printf("Environment Variables:\n")
			fmt.Printf("  PROXY_PORT       Proxy listening port (default: 8080)\n\n")
			return
		}
	}

	// Get port configuration
	port := os.Getenv("PROXY_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 微信API代理服务器 v%s\n", Version)
	fmt.Printf("监听端口: %s\n", port)
	fmt.Printf("允许的域名: %v\n", allowedDomains)

	// Start HTTP server - use Server struct for better control
	server := &http.Server{
		Addr:    ":" + port,
		Handler: http.HandlerFunc(proxyHandler),
	}
	
	log.Printf("代理服务器已启动，监听端口: %s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 