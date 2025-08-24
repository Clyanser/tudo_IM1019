package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"tudo_IM1019/common/etcd"
)

type Config struct {
	Addr      string
	Etcd      string
	Whitelist []string          `json:"whitelist"`
	Services  map[string]string `json:"services"` // service name -> etcd key suffix
	Log       logx.LogConf      `json:"log"`
}

var c Config
var configFile = flag.String("f", "setting.yaml", "the config file")

// 用于匹配 /api/{service}/ 路径
var pathRegex, _ = regexp.Compile(`/api/(.*?)/`)

// Gateway 网关结构体
type Gateway struct {
	etcdEndpoints []string
	whitelist     map[string]bool
	services      map[string]string // service name -> etcd key suffix (e.g., "auth_api")
	transport     *http.Transport
}

// NewGateway 创建网关实例
func NewGateway(config Config) *Gateway {
	whitelist := make(map[string]bool)
	for _, path := range config.Whitelist {
		whitelist[path] = true
	}

	return &Gateway{
		etcdEndpoints: strings.Split(config.Etcd, ","),
		whitelist:     whitelist,
		services:      config.Services,
		transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

// ServeHTTP 实现 http.Handler
func (g *Gateway) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	start := time.Now()

	// 1. 解析服务名
	matches := pathRegex.FindStringSubmatch(req.URL.Path)
	if len(matches) != 2 {
		http.Error(rw, "invalid request path", http.StatusBadRequest)
		return
	}
	serviceName := matches[1]

	// 2. 查找服务在配置中是否注册
	etcdKeySuffix, ok := g.services[serviceName]
	if !ok {
		logx.Errorf("service not configured: %s", serviceName)
		http.Error(rw, "service not found", http.StatusNotFound)
		return
	}

	// 3. 从 etcd 获取服务地址
	serviceAddr := etcd.GetServiceAddr(g.etcdEndpoints[0], etcdKeySuffix)
	if serviceAddr == "" {
		logx.Errorf("service not found in etcd: %s (%s)", serviceName, etcdKeySuffix)
		http.Error(rw, "service unavailable", http.StatusBadGateway)
		return
	}

	// 4. 构建目标 URL
	target, err := url.Parse(fmt.Sprintf("http://%s", serviceAddr))
	if err != nil {
		logx.Errorf("invalid target URL: %v", err)
		http.Error(rw, "internal error", http.StatusInternalServerError)
		return
	}

	// 5. 白名单检查
	if !g.whitelist[req.URL.Path] {
		if err := g.authenticate(req); err != nil {
			logx.Errorf("authentication failed: %v", err)
			rw.WriteHeader(http.StatusForbidden)
			rw.Write([]byte(`{"code":-1,"msg":"forbidden"}`))
			return
		}
	}

	// 6. 创建 ReverseProxy
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = g.transport

	// 自定义 Director，修改请求
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// 保留原始路径
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		// 添加 X-Forwarded-For
		if clientIP, _, _ := net.SplitHostPort(req.RemoteAddr); clientIP != "" {
			if prior, ok := req.Header["X-Forwarded-For"]; ok {
				clientIP = strings.Join(prior, ", ") + ", " + clientIP
			}
			req.Header.Set("X-Forwarded-For", clientIP)
		}
	}

	// 7. 执行代理
	proxy.ServeHTTP(rw, req)

	// 8. 日志
	latency := time.Since(start)
	logx.Infof("method=%s path=%s status=%s latency=%v", req.Method, req.URL.Path, rw.Header().Get("Status"), latency)
}

// authenticate 调用认证服务验证请求
func (g *Gateway) authenticate(req *http.Request) error {
	// 读取 body
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("read body failed: %v", err)
	}
	req.Body.Close()
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes)) // 恢复

	// 获取 auth 服务地址
	authAddr := etcd.GetServiceAddr(g.etcdEndpoints[0], "auth_api")
	if authAddr == "" {
		return fmt.Errorf("auth service not found in etcd")
	}

	// 构造认证请求
	authURL := fmt.Sprintf("http://%s/api/auth/authentication", authAddr)
	authReq, err := http.NewRequest("POST", authURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}

	// 拷贝 header
	authReq.Header = req.Header.Clone()

	// 添加 X-Forwarded-For
	if clientIP, _, _ := net.SplitHostPort(req.RemoteAddr); clientIP != "" {
		authReq.Header.Set("X-Forwarded-For", clientIP)
	}

	// 发起请求
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(authReq)
	if err != nil {
		return fmt.Errorf("call auth service failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read auth response failed: %v", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth service returned error: %d", resp.StatusCode)
	}

	// 解析 JSON
	var baseResp struct {
		Code int `json:"code"`
		Data *struct {
			UserID uint64 `json:"userId"`
			Role   int    `json:"role"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &baseResp); err != nil {
		return fmt.Errorf("unmarshal auth response failed: %v", err)
	}

	if baseResp.Code != 0 {
		return fmt.Errorf("auth failed: code=%d", baseResp.Code)
	}

	// 注入用户信息到 header
	if baseResp.Data != nil {
		req.Header.Set("User-ID", fmt.Sprintf("%d", baseResp.Data.UserID))
		req.Header.Set("Role", fmt.Sprintf("%d", baseResp.Data.Role))
	}

	return nil
}

func main() {
	flag.Parse()

	// 加载配置
	conf.MustLoad(*configFile, &c)

	// 设置日志
	if err := logx.SetUp(c.Log); err != nil {
		log.Fatalf("log setup failed: %v", err)
	}

	// 创建网关
	gateway := NewGateway(c)

	// 注册路由
	http.Handle("/", gateway)

	logx.Infof("Gateway is running on %s", c.Addr)
	logx.Must(http.ListenAndServe(c.Addr, nil))
}
