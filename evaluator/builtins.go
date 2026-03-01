package evaluator

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"cvb-lang/ast"
	"cvb-lang/object"
)

var builtins = map[string]*object.Builtin{
	"len": {
		Name: "len",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			default:
				return newError("argument to `len` not supported, got %s", args[0].Type())
			}
		},
	},
	"type": {
		Name: "type",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			return &object.String{Value: string(args[0].Type())}
		},
	},
	"print": {
		Name: "print",
		Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Print(arg.Inspect())
			}
			fmt.Println()
			return NULL
		},
	},
}

func applyArrayMethod(array *object.Array, method string, args []object.Object) object.Object {
	switch method {
	case "add":
		if len(args) < 1 {
			return newError("add requires at least 1 argument")
		}
		array.Elements = append(array.Elements, args...)
		return array
	case "dele":
		if len(args) != 1 {
			return newError("dele requires 1 argument")
		}
		idx, ok := args[0].(*object.Integer)
		if !ok {
			return newError("dele index must be integer")
		}
		i := idx.Value
		if i < 0 || i >= int64(len(array.Elements)) {
			return newError("index out of bounds")
		}
		array.Elements = append(array.Elements[:i], array.Elements[i+1:]...)
		return array
	case "read":
		if len(args) != 1 {
			return newError("read requires 1 argument")
		}
		idx, ok := args[0].(*object.Integer)
		if !ok {
			return newError("read index must be integer")
		}
		i := idx.Value
		if i < 0 || i >= int64(len(array.Elements)) {
			return NULL
		}
		return array.Elements[i]
	case "len":
		return &object.Integer{Value: int64(len(array.Elements))}
	default:
		return newError("unknown array method: %s", method)
	}
}

func applyStringMethod(str *object.String, method string, args []object.Object) object.Object {
	switch method {
	case "len":
		return &object.Integer{Value: int64(len(str.Value))}
	case "upper":
		return &object.String{Value: strings.ToUpper(str.Value)}
	case "lower":
		return &object.String{Value: strings.ToLower(str.Value)}
	default:
		return newError("unknown string method: %s", method)
	}
}

func applyFileMethod(file *object.File, method string, args []object.Object) object.Object {
	switch method {
	case "read":
		data, err := os.ReadFile(file.Path)
		if err != nil {
			return newError("cannot read file: %s", err.Error())
		}
		return &object.String{Value: string(data)}
	case "open":
		if len(args) < 1 {
			return newError("open requires content argument")
		}
		content, ok := args[0].(*object.String)
		if !ok {
			return newError("open content must be string")
		}
		err := os.WriteFile(file.Path, []byte(content.Value), 0644)
		if err != nil {
			return newError("cannot write file: %s", err.Error())
		}
		return NULL
	case "delete":
		err := os.Remove(file.Path)
		if err != nil {
			return newError("cannot delete file: %s", err.Error())
		}
		return NULL
	case "data":
		if len(args) < 1 {
			return newError("data requires extension argument")
		}
		ext, ok := args[0].(*object.String)
		if !ok {
			return newError("data extension must be string")
		}
		newPath := file.Path + ext.Value
		file.Path = newPath
		return file
	default:
		return newError("unknown file method: %s", method)
	}
}

func applyModuleMethod(mod *object.Module, method string, args []object.Object) object.Object {
	if val, ok := mod.Env.Get(method); ok {
		if fn, ok := val.(*object.Function); ok {
			env := createFunctionEnv(fn, args)
			result := Eval(fn.Body.(*ast.BlockStatement), env)
			return unwrapReturnValue(result)
		}
		return val
	}
	return newError("unknown module method: %s", method)
}

func createMathModule(env *object.Environment) *object.Module {
	mathEnv := object.NewEnvironment()

	mathEnv.Set("sin", &object.Builtin{
		Name: "sin",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("sin requires 1 argument")
			}
			val := toFloat(args[0])
			return &object.Float{Value: math.Sin(val)}
		},
	})

	mathEnv.Set("cos", &object.Builtin{
		Name: "cos",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("cos requires 1 argument")
			}
			val := toFloat(args[0])
			return &object.Float{Value: math.Cos(val)}
		},
	})

	mathEnv.Set("tan", &object.Builtin{
		Name: "tan",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("tan requires 1 argument")
			}
			val := toFloat(args[0])
			return &object.Float{Value: math.Tan(val)}
		},
	})

	mathEnv.Set("sqrt", &object.Builtin{
		Name: "sqrt",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("sqrt requires 1 argument")
			}
			val := toFloat(args[0])
			return &object.Float{Value: math.Sqrt(val)}
		},
	})

	mathEnv.Set("pow", &object.Builtin{
		Name: "pow",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("pow requires 2 arguments")
			}
			base := toFloat(args[0])
			exp := toFloat(args[1])
			return &object.Float{Value: math.Pow(base, exp)}
		},
	})

	mathEnv.Set("abs", &object.Builtin{
		Name: "abs",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("abs requires 1 argument")
			}
			val := toFloat(args[0])
			return &object.Float{Value: math.Abs(val)}
		},
	})

	mathEnv.Set("floor", &object.Builtin{
		Name: "floor",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("floor requires 1 argument")
			}
			val := toFloat(args[0])
			return &object.Integer{Value: int64(math.Floor(val))}
		},
	})

	mathEnv.Set("ceil", &object.Builtin{
		Name: "ceil",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("ceil requires 1 argument")
			}
			val := toFloat(args[0])
			return &object.Integer{Value: int64(math.Ceil(val))}
		},
	})

	mathEnv.Set("pi", &object.Float{Value: math.Pi})
	mathEnv.Set("e", &object.Float{Value: math.E})

	return &object.Module{Name: "math", Env: mathEnv}
}

func createRandomModule(env *object.Environment) *object.Module {
	randEnv := object.NewEnvironment()

	randEnv.Set("int", &object.Builtin{
		Name: "int",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("random.int requires 2 arguments (min, max)")
			}
			min := toInt(args[0])
			max := toInt(args[1])
			if min > max {
				return newError("min must be <= max")
			}
			return &object.Integer{Value: int64(rand.Intn(int(max-min)+1) + int(min))}
		},
	})

	randEnv.Set("float", &object.Builtin{
		Name: "float",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 0 && len(args) != 2 {
				return newError("random.float requires 0 or 2 arguments")
			}
			if len(args) == 0 {
				return &object.Float{Value: rand.Float64()}
			}
			min := toFloat(args[0])
			max := toFloat(args[1])
			return &object.Float{Value: min + rand.Float64()*(max-min)}
		},
	})

	randEnv.Set("choice", &object.Builtin{
		Name: "choice",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("random.choice requires 1 argument (array)")
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("random.choice requires array argument")
			}
			if len(arr.Elements) == 0 {
				return NULL
			}
			idx := rand.Intn(len(arr.Elements))
			return arr.Elements[idx]
		},
	})

	return &object.Module{Name: "random", Env: randEnv}
}

func createFileModule(env *object.Environment) *object.Module {
	fileEnv := object.NewEnvironment()

	fileEnv.Set("read", &object.Builtin{
		Name: "read",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("file.read requires 1 argument (file object)")
			}
			file, ok := args[0].(*object.File)
			if !ok {
				return newError("file.read requires file object")
			}
			data, err := os.ReadFile(file.Path)
			if err != nil {
				return newError("cannot read file: %s", err.Error())
			}
			return &object.String{Value: string(data)}
		},
	})

	fileEnv.Set("write", &object.Builtin{
		Name: "write",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("file.write requires 2 arguments (file object, content)")
			}
			file, ok := args[0].(*object.File)
			if !ok {
				return newError("file.write requires file object")
			}
			content, ok := args[1].(*object.String)
			if !ok {
				return newError("file.write requires string content")
			}
			err := os.WriteFile(file.Path, []byte(content.Value), 0644)
			if err != nil {
				return newError("cannot write file: %s", err.Error())
			}
			return NULL
		},
	})

	fileEnv.Set("delete", &object.Builtin{
		Name: "delete",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("file.delete requires 1 argument (file object)")
			}
			file, ok := args[0].(*object.File)
			if !ok {
				return newError("file.delete requires file object")
			}
			err := os.Remove(file.Path)
			if err != nil {
				return newError("cannot delete file: %s", err.Error())
			}
			return NULL
		},
	})

	fileEnv.Set("exists", &object.Builtin{
		Name: "exists",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("file.exists requires 1 argument (file object)")
			}
			file, ok := args[0].(*object.File)
			if !ok {
				return newError("file.exists requires file object")
			}
			_, err := os.Stat(file.Path)
			return nativeBoolToBooleanObject(!os.IsNotExist(err))
		},
	})

	return &object.Module{Name: "file", Env: fileEnv}
}

func toFloat(obj object.Object) float64 {
	switch v := obj.(type) {
	case *object.Integer:
		return float64(v.Value)
	case *object.Float:
		return v.Value
	case *object.String:
		f, _ := strconv.ParseFloat(v.Value, 64)
		return f
	default:
		return 0
	}
}

func toInt(obj object.Object) int64 {
	switch v := obj.(type) {
	case *object.Integer:
		return v.Value
	case *object.Float:
		return int64(v.Value)
	case *object.String:
		i, _ := strconv.ParseInt(v.Value, 10, 64)
		return i
	default:
		return 0
	}
}
