package main

//
//import (
//	"bytes"
//	"encoding/json"
//	"flag"
//	"fmt"
//	"github.com/zeromicro/go-zero/core/conf"
//	"github.com/zeromicro/go-zero/core/logx"
//	"io"
//	"net/http"
//	"regexp"
//	"strings"
//	"time"
//
//	"tudo_IM1019/common/etcd"
//)
//
//type BaseResponse struct {
//	Code int    `json:"code"`
//	Msg  string `json:"msg"`
//	Data any    `json:"data"`
//}
//
//func FailResponse(msg string, res http.ResponseWriter) {
//	response := BaseResponse{Code: -1, Msg: msg}
//	byteData, _ := json.Marshal(response)
//	res.Write(byteData)
//}
//
//// 服务地址映射（备用）
//var IM_ServiceMap = map[string]string{
//	"auth":  "http://127.0.0.1:20021",
//	"user":  "http://127.0.0.1:20022",
//	"chat":  "http://127.0.0.1:20023",
//	"group": "http://127.0.0.1:20024",
//}
//
//type Config struct {
//	Addr string
//	Etcd string
//	log  logx.LogConf
//}
//
//var c Config
//var configFile = flag.String("f", "settings.yaml", "the config file")
//
//func IM_gateway(res http.ResponseWriter, req *http.Request) {
//	//路由解析，识别请求的目标服务
//	// 匹配请求前缀 /api/user/xx
//	regex, _ := regexp.Compile(`/api/(.*?)/`)
//	addList := regex.FindStringSubmatch(req.URL.Path)
//	if len(addList) != 2 {
//		http.Error(res, "error", http.StatusBadRequest)
//		return
//	}
//	//服务发现
//	// 获取目标服务名
//	service := addList[1]
//	addr := etcd.GetServiceAddr(c.Etcd, service+"_api")
//	if addr == "" {
//		logx.Errorf("未找到对应的服务地址: %s", service+"_api")
//		http.Error(res, "服务不可用", http.StatusBadGateway)
//		return
//	}
//	//读取请求体并恢复
//	// 读取 body 内容（保存为 byte[]）
//	bodyBytes, err := io.ReadAll(req.Body)
//	if err != nil {
//		http.Error(res, "读取请求体失败", http.StatusInternalServerError)
//		return
//	}
//	req.Body.Close() // 关闭原始 body
//
//	// 恢复 body，供后续 proxyReq 使用
//	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
//	//调试信息
//	//logx.Infof("bodybytes %s", string(bodyBytes))  // 查看内容
//	//logx.Infof("bodybytes 长度: %d", len(bodyBytes)) // 查看长度
//	// 获取客户端 IP
//	// 添加 X-Forwarded-For
//	remoteIP := strings.Split(req.RemoteAddr, ":")[0]
//
//	// 白名单机制（避免登录接口被拦截）
//	whiteList := map[string]bool{
//		"/api/auth/login":          true,
//		"/api/auth/register":       true,
//		"/api/auth/logout":         true,
//		"/api/auth/authentication": true,
//	}
//	if _, ok := whiteList[req.URL.Path]; !ok {
//		// 请求认证服务的地址
//		authAddr := etcd.GetServiceAddr(c.Etcd, "auth_api")
//		if authAddr == "" {
//			logx.Errorf("认证服务未注册到 etcd")
//			res.WriteHeader(http.StatusUnauthorized)
//			FailResponse("认证服务不可用", res)
//			return
//		}
//
//		authUrl := fmt.Sprintf("http://%s/api/auth/authentication", authAddr)
//		logx.Infof("构造的认证服务 URL：%s", authUrl)
//		//调试信息
//		//logx.Infof("bodybytes %s", string(bodyBytes))  // 查看内容
//		//logx.Infof("bodybytes 长度: %d", len(bodyBytes)) // 查看长度
//		authReq, _ := http.NewRequest("POST", authUrl, bytes.NewReader(bodyBytes))
//
//		//拷贝请求头
//		for k, vv := range req.Header {
//			for _, v := range vv {
//				authReq.Header.Add(k, v)
//			}
//		}
//
//		authReq.Header.Set("X-Forwarded-For", remoteIP)
//
//		client := &http.Client{Timeout: 5 * time.Second} // 带超时的 client
//		authResp, err := client.Do(authReq)
//		if err != nil {
//			logx.Errorf("认证服务调用失败: %v", err)
//			res.WriteHeader(http.StatusUnauthorized)
//			FailResponse("认证服务不可用", res)
//			return
//		}
//		defer authResp.Body.Close()
//
//		type Resp struct {
//			Code int    `json:"code"`
//			Msg  string `json:"msg"`
//			Data *struct {
//				UserID uint64 `json:"userId"`
//				Role   int    `json:"role"`
//			} `json:"data"`
//		}
//
//		bytedata, readErr := io.ReadAll(authResp.Body)
//		if readErr != nil {
//			logx.Errorf("读取认证服务响应失败: %v", readErr)
//			res.WriteHeader(http.StatusInternalServerError)
//			FailResponse("认证服务不可用", res)
//			return
//		}
//
//		// 判断是否是有效的 JSON
//		if !json.Valid(bytedata) {
//			logx.Errorf("认证服务返回非 JSON 数据: %s", string(bytedata))
//			res.WriteHeader(http.StatusInternalServerError)
//			FailResponse("认证服务返回格式错误", res)
//			return
//		}
//
//		var AuthResponse Resp
//		autherr := json.Unmarshal(bytedata, &AuthResponse)
//		if autherr != nil {
//			logx.Errorf("JSON解析失败: %v, 原始数据: %s", autherr, string(bytedata))
//			res.WriteHeader(http.StatusInternalServerError)
//			FailResponse("认证服务返回格式错误", res)
//			return
//		}
//
//		// 认证不通过
//		if AuthResponse.Code != 0 {
//			res.WriteHeader(http.StatusForbidden)
//			res.Write(bytedata)
//			return
//		}
//		//调试信息
//		logx.Infof("Setting headers -> User-ID: %d, Role: %d", AuthResponse.Data.UserID, AuthResponse.Data.Role)
//		//传递用户信息
//		if AuthResponse.Data != nil {
//			req.Header.Set("User-ID", fmt.Sprintf("%d", AuthResponse.Data.UserID))
//			req.Header.Set("Role", fmt.Sprintf("%d", AuthResponse.Data.Role))
//		}
//	}
//	//构造代理请求（反向代理）
//	// 构建新的请求
//	url := fmt.Sprintf("http://%s%s", addr, req.URL.String())
//	proxyReq, err := http.NewRequest(req.Method, url, bytes.NewReader(bodyBytes))
//	if err != nil {
//		http.Error(res, "创建代理请求失败", http.StatusInternalServerError)
//		return
//	}
//
//	// 拷贝原始请求头
//	for k, vv := range req.Header {
//		for _, v := range vv {
//			proxyReq.Header.Add(k, v)
//		}
//	}
//
//	proxyReq.Header.Set("X-Forwarded-For", remoteIP)
//
//	// 发起代理请求
//	client := &http.Client{Timeout: 10 * time.Second}
//	response, err := client.Do(proxyReq)
//	if err != nil {
//		logx.Errorf("请求服务异常: %v", err)
//		http.Error(res, "服务异常", http.StatusBadGateway)
//		return
//	}
//	defer response.Body.Close()
//
//	// 拷贝响应头
//	for k, vv := range response.Header {
//		for _, v := range vv {
//			res.Header().Add(k, v)
//		}
//	}
//
//	// 设置状态码
//	res.WriteHeader(response.StatusCode)
//
//	// 拷贝响应体
//	io.Copy(res, response.Body)
//
//	logx.Infof("请求路径：%s -> %s", req.URL.Path, url)
//}
//
//func main() {
//	flag.Parse()
//	conf.MustLoad(*configFile, &c)
//
//	err := logx.SetUp(c.log)
//	if err != nil {
//		return
//	}
//
//	// 注册回调函数
//	http.HandleFunc("/", IM_gateway)
//	fmt.Printf("网关运行在 %s\n", c.Addr)
//
//	// 启动服务
//	logx.Must(http.ListenAndServe(c.Addr, nil))
//}
//
////总结：
////客户端请求 → 网关收到请求 → 解析路径 → 提取服务名 → 从 etcd 获取服务地址
////→ 读取 body 并恢复 → 判断是否白名单接口
////→ 否：调用认证服务 → 成功才继续
////→ 是：构造代理请求 → 发送给目标服务
////→ 获取响应 → 返回客户端
////→ 记录日志
