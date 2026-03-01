# CVB 语言 Go 实现 🚀

这是一个用 Go 语言实现的 CVB 编程语言解释器。

官网:https://keen-mandazi-dfe178.netlify.app/
(AI开发的官网哈！）

## 项目结构 📁

```
cvb-lang/
├── main.go              # 主程序入口
├── lexer/               # 词法分析器
│   └── lexer.go
├── parser/              # 语法分析器
│   └── parser.go
├── ast/                 # 抽象语法树
│   └── ast.go
├── object/              # 对象系统
│   └── object.go
├── evaluator/           # 求值器
│   ├── evaluator.go
│   └── builtins.go
├── go.mod               # Go 模块文件
├── test.cvb             # 测试示例
└── README.md            # 本文件
```

## 已实现的功能 ✅

### 1. 基础语法
- ✅ 变量定义: `name=>str&"value"`
- ✅ 输出: `#print=>str&"hello"`
- ✅ 注释: `(# comment #)`

### 2. 数据类型
- ✅ 字符串 (str)
- ✅ 整数 (int)
- ✅ 浮点数 (float)
- ✅ 布尔值 (TRUE/FALSE)
- ✅ 列表 (list): `["a", "b", "c"]`
- ✅ 字典 (dic): `{"key": "value"}`

### 3. 控制流
- ✅ if/else 条件判断
- ✅ while 循环 (支持条件)
- ✅ for 循环 (遍历列表/字符串)
- ✅ break 语句

### 4. 函数
- ✅ 函数定义: `#def name(): { body }`
- ✅ 函数调用
- ✅ 参数传递

### 5. 内置模块
- ✅ **file 模块**: 文件读写、删除、检查存在
- ✅ **math 模块**: sin, cos, tan, sqrt, pow, abs, floor, ceil, pi, e
- ✅ **random 模块**: int, float, choice
- ✅ **net 模块**: HTTP 请求、服务器创建、服务响应
  - `net.get(url)` - GET 请求
  - `net.post(url, data)` - POST 请求
  - `net.request(config)` - 通用 HTTP 请求
  - `net.sever(config)` - 创建 HTTP 服务器
  - `net.severgo(config)` - 创建动态响应服务器
- ✅ **shell 模块**: 系统命令执行、环境变量操作
  - `shell.exec(command)` - 执行系统命令
  - `shell.run(command)` - 执行命令并返回详细结果
  - `shell.output(command)` - 只获取命令输出
  - `shell.system(command)` - 直接输出到控制台
  - `shell.pwd()` - 获取当前目录
  - `shell.cd(path)` - 切换目录
  - `shell.getenv(name)` - 获取环境变量
  - `shell.setenv(name, value)` - 设置环境变量
  - `shell.which(command)` - 查找命令路径

### 6. 运算符
- ✅ 算术: +, -, *, /, %
- ✅ 比较: ==, !=, <, >, <=, >=
- ✅ 逻辑: &&, ||, and, or
- ✅ 索引: `array[index]`
- ✅ 方法调用: `list.add<"item">`

## 使用方法 📖

### 编译
```bash
go build -o cvb.exe .
```

### 运行 REPL
```bash
./cvb.exe
```

### 运行文件
```bash
./cvb.exe test.cvb
```

## 示例代码 💡

```cvb
(# 变量定义 #)
name=>str&"Hello CVB!"
age=>int&25

(# 输出 #)
#print=>str&name

(# 列表操作 #)
mylist=>list&["apple", "banana"]
mylist.add<"orange">

(# if 判断 #)
#if age >= 18:
  #print=>str&"成年人"
#else:
  #print=>str&"未成年人"

(# while 循环 #)
#while TRUE=5:
  #print=>str&"循环中"

(# for 循环 #)
#for item in mylist:
  #print=>str&item

(# 引入模块 #)
#import<file, math, random, net, shell>

(# 数学运算 #)
result=>math.sqrt(16)

(# 随机数 #)
rand_num=>random.int(1, 100)

(# HTTP 请求 #)
response=>net.get("https://api.github.com")
#print=>int&response.status

(# HTTP 服务器 #)
server_config=>dic&{"port": 8080, "type": "http"}
#server=>net.sever(server_config)

(# 执行系统命令 #)
#cmd_result=>shell.exec("echo Hello World")
#if cmd_result.success:
  #print=>str&cmd_result.stdout

(# 获取当前目录 #)
#current_dir=>shell.pwd()
#print=>str&current_dir
```

## 技术栈 🛠️

- **Go 1.18+** - 编程语言
- 纯 Go 实现，无外部依赖
## 注意事项 ⚠️

1. 需要安装 Go 1.18 或更高版本
2. Windows 系统直接运行 `cvb.exe`
3. 文件路径使用正斜杠 `/` 或双反斜杠 `\\`
```说明
   （可能更新就越来越少，但不会断更，只要有star就有更下去的动力，因为本人要上学）
```
