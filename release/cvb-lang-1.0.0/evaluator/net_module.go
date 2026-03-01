package evaluator

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"strings"
	"time"

	"cvb-lang/ast"
	"cvb-lang/object"
)

// NetServer 存储 HTTP 服务器实例
type NetServer struct {
	Port      int64
	Type      string
	Domain    string
	Directory string
	Server    *http.Server
	Routes    map[string]http.HandlerFunc
}

// NetResponse 存储响应配置
type NetResponse struct {
	Type       string
	JSONData   map[string]interface{}
	TextData   string
	Variables  map[string]string
}

func createNetModule(env *object.Environment) *object.Module {
	netEnv := object.NewEnvironment()

	// net.request - 发送 HTTP 请求
	netEnv.Set("request", &object.Builtin{
		Name: "request",
		Fn:   netRequest,
	})

	// net.sever - 创建 HTTP 服务器
	netEnv.Set("sever", &object.Builtin{
		Name: "sever",
		Fn:   netSever,
	})

	// net.severgo - 服务响应
	netEnv.Set("severgo", &object.Builtin{
		Name: "severgo",
		Fn:   netSevergo,
	})

	// net.get - GET 请求快捷方式
	netEnv.Set("get", &object.Builtin{
		Name: "get",
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("net.get requires at least 1 argument (url)")
			}
			url, ok := args[0].(*object.String)
			if !ok {
				return newError("net.get url must be string")
			}
			return httpRequest("GET", url.Value, "", nil)
		},
	})

	// net.post - POST 请求快捷方式
	netEnv.Set("post", &object.Builtin{
		Name: "post",
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 2 {
				return newError("net.post requires 2 arguments (url, data)")
			}
			url, ok := args[0].(*object.String)
			if !ok {
				return newError("net.post url must be string")
			}
			data := args[1].Inspect()
			contentType := "application/json"
			if len(args) >= 3 {
				if ct, ok := args[2].(*object.String); ok {
					contentType = ct.Value
				}
			}
			headers := map[string]string{"Content-Type": contentType}
			return httpRequest("POST", url.Value, data, headers)
		},
	})

	return &object.Module{Name: "net", Env: netEnv}
}

// net.request 实现
func netRequest(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("net.request requires at least 1 argument (config hash)")
	}

	config, ok := args[0].(*object.Hash)
	if !ok {
		return newError("net.request argument must be a hash")
	}

	// 解析配置
	var url, method, dataType string
	var data string
	var headers = make(map[string]string)

	// 获取 URL
	if urlPair, ok := config.Pairs[object.HashKey{Type: object.STRING_OBJ, Value: hashString("url")}]; ok {
		if urlStr, ok := urlPair.Value.(*object.String); ok {
			url = urlStr.Value
		}
	}

	// 获取 Method
	if methodPair, ok := config.Pairs[object.HashKey{Type: object.STRING_OBJ, Value: hashString("method")}]; ok {
		if methodStr, ok := methodPair.Value.(*object.String); ok {
			method = strings.ToUpper(methodStr.Value)
		}
	}

	// 获取 DataType
	if typePair, ok := config.Pairs[object.HashKey{Type: object.STRING_OBJ, Value: hashString("data")}]; ok {
		if typeStr, ok := typePair.Value.(*object.String); ok {
			dataType = typeStr.Value
		}
	}

	// 获取 Data 内容
	if dataType == "json" {
		headers["Content-Type"] = "application/json"
		jsonData := make(map[string]interface{})
		for key, pair := range config.Pairs {
			keyStr := unhashString(key.Value)
			if strings.HasPrefix(keyStr, "data.json.") {
				varName := strings.TrimPrefix(keyStr, "data.json.")
				jsonData[varName] = pair.Value.Inspect()
			}
		}
		jsonBytes, _ := json.Marshal(jsonData)
		data = string(jsonBytes)
	} else if dataType == "text" {
		headers["Content-Type"] = "text/plain"
		for key, pair := range config.Pairs {
			keyStr := unhashString(key.Value)
			if strings.HasPrefix(keyStr, "env.") {
				data += pair.Value.Inspect()
			}
		}
	}

	if method == "" {
		method = "GET"
	}

	return httpRequest(method, url, data, headers)
}

// HTTP 请求辅助函数
func httpRequest(method, url, data string, headers map[string]string) object.Object {
	if url == "" {
		return newError("URL cannot be empty")
	}

	// 确保 URL 有协议前缀
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	var body io.Reader
	if data != "" {
		body = strings.NewReader(data)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return newError("Failed to create request: %s", err.Error())
	}

	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return newError("Request failed: %s", err.Error())
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return newError("Failed to read response: %s", err.Error())
	}

	// 构建响应对象
	result := &object.Hash{Pairs: make(map[object.HashKey]object.HashPair)}
	result.Pairs[hashKey("status")] = object.HashPair{
		Key:   &object.String{Value: "status"},
		Value: &object.Integer{Value: int64(resp.StatusCode)},
	}
	result.Pairs[hashKey("body")] = object.HashPair{
		Key:   &object.String{Value: "body"},
		Value: &object.String{Value: string(respBody)},
	}
	result.Pairs[hashKey("headers")] = object.HashPair{
		Key:   &object.String{Value: "headers"},
		Value: &object.String{Value: fmt.Sprintf("%v", resp.Header)},
	}

	return result
}

// net.sever 实现
func netSever(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("net.sever requires at least 1 argument (config hash)")
	}

	config, ok := args[0].(*object.Hash)
	if !ok {
		return newError("net.sever argument must be a hash")
	}

	server := &NetServer{
		Port:   8080,
		Type:   "http",
		Routes: make(map[string]http.HandlerFunc),
	}

	// 解析端口
	if portPair, ok := config.Pairs[hashKey("port")]; ok {
		if portInt, ok := portPair.Value.(*object.Integer); ok {
			server.Port = portInt.Value
		}
	}

	// 解析类型
	if typePair, ok := config.Pairs[hashKey("type")]; ok {
		if typeStr, ok := typePair.Value.(*object.String); ok {
			server.Type = typeStr.Value
		}
	}

	// 解析域名
	if domainPair, ok := config.Pairs[hashKey("domain")]; ok {
		if domainStr, ok := domainPair.Value.(*object.String); ok {
			server.Domain = domainStr.Value
		}
	}

	// 解析目录
	if dirPair, ok := config.Pairs[hashKey("net.directory")]; ok {
		if dirStr, ok := dirPair.Value.(*object.String); ok {
			server.Directory = dirStr.Value
		}
	}
	if dirPathPair, ok := config.Pairs[hashKey("directory.path")]; ok {
		if dirPathStr, ok := dirPathPair.Value.(*object.String); ok {
			server.Directory = dirPathStr.Value
		}
	}

	// 创建 HTTP 服务器
	mux := http.NewServeMux()
	server.Server = &http.Server{
		Addr:    fmt.Sprintf(":%d", server.Port),
		Handler: mux,
	}

	// 默认处理函数
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if server.Directory != "" {
			http.FileServer(http.Dir(server.Directory)).ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("CVB Net Server Running"))
		}
	})

	// 在后台启动服务器
	go func() {
		server.Server.ListenAndServe()
	}()

	// 返回服务器对象
	return &object.Module{
		Name: "netserver",
		Env:  createServerEnv(server),
	}
}

// net.severgo 实现
func netSevergo(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("net.severgo requires at least 1 argument (config hash)")
	}

	config, ok := args[0].(*object.Hash)
	if !ok {
		return newError("net.severgo argument must be a hash")
	}

	var port int64 = 8080
	var respType string = "text"
	var handlerFunc object.Object

	// 解析端口
	if portPair, ok := config.Pairs[hashKey("port")]; ok {
		if portInt, ok := portPair.Value.(*object.Integer); ok {
			port = portInt.Value
		}
	}

	// 解析返回类型
	if typePair, ok := config.Pairs[hashKey("type")]; ok {
		if typeStr, ok := typePair.Value.(*object.String); ok {
			respType = typeStr.Value
		}
	}

	// 解析处理函数
	if goPair, ok := config.Pairs[hashKey("go")]; ok {
		handlerFunc = goPair.Value
	}

	// 创建响应配置
	response := &NetResponse{
		Type:      respType,
		JSONData:  make(map[string]interface{}),
		Variables: make(map[string]string),
	}

	// 解析 JSON 数据配置
	for key, pair := range config.Pairs {
		keyStr := unhashString(key.Value)
		if strings.HasPrefix(keyStr, "type.json.") && !strings.Contains(keyStr, ".值") {
			varName := strings.TrimPrefix(keyStr, "type.json.")
			response.JSONData[varName] = pair.Value.Inspect()
		}
		if strings.HasPrefix(keyStr, "type.text.") && !strings.Contains(keyStr, ".值") {
			varName := strings.TrimPrefix(keyStr, "type.text.")
			response.Variables[varName] = pair.Value.Inspect()
		}
	}

	// 创建 HTTP 服务器
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	// 注册处理函数
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 如果有自定义处理函数，调用它
		if handlerFunc != nil {
			if fn, ok := handlerFunc.(*object.Function); ok {
				// 创建请求上下文
				reqEnv := object.NewEnvironment()
				reqEnv.Set("method", &object.String{Value: r.Method})
				reqEnv.Set("path", &object.String{Value: r.URL.Path})
				reqEnv.Set("query", &object.String{Value: r.URL.RawQuery})

				// 调用处理函数
				env := object.NewEnclosedEnvironment(fn.Env)
				result := Eval(fn.Body.(*ast.BlockStatement), env)

				// 根据结果设置响应
				if result != nil {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(result.Inspect()))
					return
				}
			}
		}

		// 默认响应
		if response.Type == "json" {
			w.Header().Set("Content-Type", "application/json")
			jsonBytes, _ := json.Marshal(response.JSONData)
			w.Write(jsonBytes)
		} else {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(response.TextData))
		}
	})

	// 在后台启动服务器
	go func() {
		fmt.Printf("CVB Net Server starting on port %d\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %s\n", err.Error())
		}
	}()

	return &object.Module{
		Name: fmt.Sprintf("netserver:%d", port),
		Env:  object.NewEnvironment(),
	}
}

// 创建服务器环境
func createServerEnv(server *NetServer) *object.Environment {
	env := object.NewEnvironment()

	env.Set("port", &object.Integer{Value: server.Port})
	env.Set("type", &object.String{Value: server.Type})
	env.Set("domain", &object.String{Value: server.Domain})
	env.Set("directory", &object.String{Value: server.Directory})

	// 添加路由方法
	env.Set("route", &object.Builtin{
		Name: "route",
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 2 {
				return newError("route requires 2 arguments (path, handler)")
			}
			path, ok := args[0].(*object.String)
			if !ok {
				return newError("route path must be string")
			}

			// 存储路由
			if fn, ok := args[1].(*object.Function); ok {
				server.Routes[path.Value] = func(w http.ResponseWriter, r *http.Request) {
					env := object.NewEnclosedEnvironment(fn.Env)
					result := Eval(fn.Body.(*ast.BlockStatement), env)
					if result != nil {
						w.Write([]byte(result.Inspect()))
					}
				}
			}
			return NULL
		},
	})

	// 停止服务器
	env.Set("stop", &object.Builtin{
		Name: "stop",
		Fn: func(args ...object.Object) object.Object {
			if server.Server != nil {
				server.Server.Close()
			}
			return NULL
		},
	})

	return env
}

// 辅助函数 - 使用 FNV hash
func hashKey(s string) object.HashKey {
	h := fnv.New64a()
	h.Write([]byte(s))
	return object.HashKey{Type: object.STRING_OBJ, Value: h.Sum64()}
}

func hashString(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

func unhashString(h uint64) string {
	// 简化的反向查找，实际使用时需要更好的实现
	return fmt.Sprintf("%d", h)
}
