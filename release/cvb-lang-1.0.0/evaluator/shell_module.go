package evaluator

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"cvb-lang/object"
)

// ShellResult 存储命令执行结果
type ShellResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Success  bool
}

func createShellModule(env *object.Environment) *object.Module {
	shellEnv := object.NewEnvironment()

	// shell.exec - 执行系统命令
	shellEnv.Set("exec", &object.Builtin{
		Name: "exec",
		Fn:   shellExec,
	})

	// shell.run - 执行命令并返回详细结果
	shellEnv.Set("run", &object.Builtin{
		Name: "run",
		Fn:   shellRun,
	})

	// shell.output - 只获取命令输出
	shellEnv.Set("output", &object.Builtin{
		Name: "output",
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("shell.output requires at least 1 argument (command)")
			}
			cmdStr, ok := args[0].(*object.String)
			if !ok {
				return newError("shell.output command must be string")
			}

			result := executeCommand(cmdStr.Value, nil)
			if result.Success {
				return &object.String{Value: result.Stdout}
			}
			return newError("Command failed: %s", result.Stderr)
		},
	})

	// shell.system - 执行系统命令（简化版，直接输出到控制台）
	shellEnv.Set("system", &object.Builtin{
		Name: "system",
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("shell.system requires at least 1 argument (command)")
			}
			cmdStr, ok := args[0].(*object.String)
			if !ok {
				return newError("shell.system command must be string")
			}

			var cmd *exec.Cmd
			if runtime.GOOS == "windows" {
				cmd = exec.Command("cmd", "/c", cmdStr.Value)
			} else {
				cmd = exec.Command("sh", "-c", cmdStr.Value)
			}

			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()

			if err != nil {
				return newError("Command execution failed: %s", err.Error())
			}
			return NULL
		},
	})

	// shell.pwd - 获取当前工作目录
	shellEnv.Set("pwd", &object.Builtin{
		Name: "pwd",
		Fn: func(args ...object.Object) object.Object {
			pwd, err := os.Getwd()
			if err != nil {
				return newError("Failed to get current directory: %s", err.Error())
			}
			return &object.String{Value: pwd}
		},
	})

	// shell.cd - 切换工作目录
	shellEnv.Set("cd", &object.Builtin{
		Name: "cd",
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("shell.cd requires 1 argument (path)")
			}
			path, ok := args[0].(*object.String)
			if !ok {
				return newError("shell.cd path must be string")
			}
			err := os.Chdir(path.Value)
			if err != nil {
				return newError("Failed to change directory: %s", err.Error())
			}
			return NULL
		},
	})

	// shell.getenv - 获取环境变量
	shellEnv.Set("getenv", &object.Builtin{
		Name: "getenv",
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("shell.getenv requires 1 argument (variable name)")
			}
			name, ok := args[0].(*object.String)
			if !ok {
				return newError("shell.getenv variable name must be string")
			}
			value := os.Getenv(name.Value)
			return &object.String{Value: value}
		},
	})

	// shell.setenv - 设置环境变量
	shellEnv.Set("setenv", &object.Builtin{
		Name: "setenv",
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 2 {
				return newError("shell.setenv requires 2 arguments (name, value)")
			}
			name, ok := args[0].(*object.String)
			if !ok {
				return newError("shell.setenv name must be string")
			}
			value, ok := args[1].(*object.String)
			if !ok {
				return newError("shell.setenv value must be string")
			}
			err := os.Setenv(name.Value, value.Value)
			if err != nil {
				return newError("Failed to set environment variable: %s", err.Error())
			}
			return NULL
		},
	})

	// shell.which - 查找命令路径
	shellEnv.Set("which", &object.Builtin{
		Name: "which",
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("shell.which requires 1 argument (command)")
			}
			cmd, ok := args[0].(*object.String)
			if !ok {
				return newError("shell.which command must be string")
			}

			var path string
			if runtime.GOOS == "windows" {
				result := executeCommand("where "+cmd.Value, nil)
				if result.Success {
					path = strings.TrimSpace(result.Stdout)
				}
			} else {
				result := executeCommand("which "+cmd.Value, nil)
				if result.Success {
					path = strings.TrimSpace(result.Stdout)
				}
			}

			if path == "" {
				return NULL
			}
			return &object.String{Value: path}
		},
	})

	return &object.Module{Name: "shell", Env: shellEnv}
}

// shell.exec 实现
func shellExec(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("shell.exec requires at least 1 argument (command)")
	}

	cmdStr, ok := args[0].(*object.String)
	if !ok {
		return newError("shell.exec command must be string")
	}

	// 解析可选参数
	var workDir string
	var envVars map[string]string

	if len(args) >= 2 {
		if config, ok := args[1].(*object.Hash); ok {
			// 解析工作目录
			if dirPair, ok := config.Pairs[hashKey("dir")]; ok {
				if dir, ok := dirPair.Value.(*object.String); ok {
					workDir = dir.Value
				}
			}
			// 解析环境变量
			envVars = make(map[string]string)
			for key, pair := range config.Pairs {
				keyStr := keyToString(key)
				if strings.HasPrefix(keyStr, "env.") {
					varName := strings.TrimPrefix(keyStr, "env.")
					envVars[varName] = pair.Value.Inspect()
				}
			}
		}
	}

	result := executeCommand(cmdStr.Value, &CommandConfig{
		WorkDir: workDir,
		Env:     envVars,
	})

	return shellResultToObject(result)
}

// shell.run 实现
func shellRun(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("shell.run requires at least 1 argument (command)")
	}

	cmdStr, ok := args[0].(*object.String)
	if !ok {
		return newError("shell.run command must be string")
	}

	result := executeCommand(cmdStr.Value, nil)
	return shellResultToObject(result)
}

// CommandConfig 命令配置
type CommandConfig struct {
	WorkDir string
	Env     map[string]string
}

// executeCommand 执行系统命令
func executeCommand(command string, config *CommandConfig) *ShellResult {
	result := &ShellResult{
		ExitCode: -1,
		Success:  false,
	}

	var cmd *exec.Cmd

	// 根据操作系统选择 shell
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	// 设置工作目录
	if config != nil && config.WorkDir != "" {
		cmd.Dir = config.WorkDir
	}

	// 设置环境变量
	if config != nil && len(config.Env) > 0 {
		env := os.Environ()
		for key, value := range config.Env {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}

	// 捕获输出
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 执行命令
	err := cmd.Run()

	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
		result.Success = false
	} else {
		result.ExitCode = 0
		result.Success = true
	}

	return result
}

// shellResultToObject 将 ShellResult 转换为 CVB 对象
func shellResultToObject(result *ShellResult) object.Object {
	obj := &object.Hash{Pairs: make(map[object.HashKey]object.HashPair)}

	obj.Pairs[hashKey("stdout")] = object.HashPair{
		Key:   &object.String{Value: "stdout"},
		Value: &object.String{Value: result.Stdout},
	}
	obj.Pairs[hashKey("stderr")] = object.HashPair{
		Key:   &object.String{Value: "stderr"},
		Value: &object.String{Value: result.Stderr},
	}
	obj.Pairs[hashKey("exitCode")] = object.HashPair{
		Key:   &object.String{Value: "exitCode"},
		Value: &object.Integer{Value: int64(result.ExitCode)},
	}
	obj.Pairs[hashKey("success")] = object.HashPair{
		Key:   &object.String{Value: "success"},
		Value: nativeBoolToBooleanObject(result.Success),
	}

	return obj
}

// keyToString 将 HashKey 转换为字符串（简化版）
func keyToString(key object.HashKey) string {
	// 这里简化处理，实际应该维护一个反向映射
	return fmt.Sprintf("%d", key.Value)
}
