package evaluator

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cvb-lang/ast"
	"cvb-lang/lexer"
	"cvb-lang/object"
	"cvb-lang/parser"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, env)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.ImportStatement:
		return evalImportStatement(node, env)
	case *ast.ImportFileStatement:
		return evalImportFileStatement(node, env)
	case *ast.PrintStatement:
		return evalPrintStatement(node, env)
	case *ast.VariableStatement:
		return evalVariableStatement(node, env)
	case *ast.WhileStatement:
		return evalWhileStatement(node, env)
	case *ast.ForStatement:
		return evalForStatement(node, env)
	case *ast.IfStatement:
		return evalIfStatement(node, env)
	case *ast.FunctionDefinition:
		return evalFunctionDefinition(node, env)
	case *ast.BreakStatement:
		return &object.ReturnValue{Value: &object.String{Value: "__BREAK__"}}
	case *ast.ReturnValue:
		val := Eval(node.Value, env)
		return &object.ReturnValue{Value: val}

	case *ast.InfixExpression:
		return evalInfixExpression(node, env)
	case *ast.PrefixExpression:
		return evalPrefixExpression(node, env)
	case *ast.CallExpression:
		return evalCallExpression(node, env)
	case *ast.MethodCallExpression:
		return evalMethodCallExpression(node, env)
	case *ast.IndexExpression:
		return evalIndexExpression(node, env)
	case *ast.PipeExpression:
		return evalPipeExpression(node, env)

	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.NumberLiteral:
		if float64(int64(node.Value)) == node.Value {
			return &object.Integer{Value: int64(node.Value)}
		}
		return &object.Float{Value: node.Value}
	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.ListLiteral:
		return evalListLiteral(node, env)
	case *ast.DictLiteral:
		return evalDictLiteral(node, env)
	}

	return newError("unknown node type: %T", node)
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func evalImportStatement(node *ast.ImportStatement, env *object.Environment) object.Object {
	for _, module := range node.Modules {
		switch module {
		case "file":
			env.Set("file", createFileModule(env))
		case "math":
			env.Set("math", createMathModule(env))
		case "random":
			env.Set("random", createRandomModule(env))
		case "net":
			env.Set("net", createNetModule(env))
		case "shell":
			env.Set("shell", createShellModule(env))
		default:
			return newError("unknown module: %s", module)
		}
	}
	return NULL
}

func evalImportFileStatement(node *ast.ImportFileStatement, env *object.Environment) object.Object {
	path := Eval(node.Path, env)
	if isError(path) {
		return path
	}

	name := Eval(node.Name, env)
	if isError(name) {
		return name
	}

	pathStr, ok := path.(*object.String)
	if !ok {
		return newError("file path must be a string, got %s", path.Type())
	}

	nameStr, ok := name.(*object.String)
	if !ok {
		return newError("file name must be a string, got %s", name.Type())
	}

	fullPath := filepath.Join(pathStr.Value, nameStr.Value)
	fileObj := &object.File{
		Path: fullPath,
	}

	env.Set(node.VarName, fileObj)
	return NULL
}

func evalPrintStatement(node *ast.PrintStatement, env *object.Environment) object.Object {
	val := Eval(node.Value, env)
	if isError(val) {
		return val
	}

	switch node.TypeHint {
	case "math":
		if str, ok := val.(*object.String); ok {
			result := evalMathExpression(str.Value)
			fmt.Println(result)
			return NULL
		}
		fmt.Println(val.Inspect())
	case "str":
		fmt.Println(val.Inspect())
	case "int":
		fmt.Println(val.Inspect())
	default:
		fmt.Println(val.Inspect())
	}

	return NULL
}

func evalVariableStatement(node *ast.VariableStatement, env *object.Environment) object.Object {
	val := Eval(node.Value, env)
	if isError(val) {
		return val
	}

	switch node.TypeHint {
	case "str":
		if _, ok := val.(*object.String); !ok {
			val = &object.String{Value: val.Inspect()}
		}
	case "int":
		switch v := val.(type) {
		case *object.Integer:
			// already int
		case *object.Float:
			val = &object.Integer{Value: int64(v.Value)}
		case *object.String:
			i, err := strconv.ParseInt(v.Value, 10, 64)
			if err == nil {
				val = &object.Integer{Value: i}
			}
		}
	case "list":
		if _, ok := val.(*object.Array); !ok {
			val = &object.Array{Elements: []object.Object{val}}
		}
	case "dic":
		if _, ok := val.(*object.Hash); !ok {
			hash := &object.Hash{Pairs: make(map[object.HashKey]object.HashPair)}
			hash.Pairs[object.HashKey{Type: object.STRING_OBJ, Value: 0}] = object.HashPair{
				Key:   &object.String{Value: "value"},
				Value: val,
			}
			val = hash
		}
	}

	env.Set(node.Name, val)
	return NULL
}

func evalWhileStatement(node *ast.WhileStatement, env *object.Environment) object.Object {
	var result object.Object = NULL

	count := int64(-1)
	if node.Count != nil {
		countVal := Eval(node.Count, env)
		if isError(countVal) {
			return countVal
		}
		if i, ok := countVal.(*object.Integer); ok {
			count = i.Value
		}
	}

	iterations := int64(0)
	for {
		if count >= 0 && iterations >= count {
			break
		}

		if node.Condition != nil {
			cond := Eval(node.Condition, env)
			if isError(cond) {
				return cond
			}
			if !isTruthy(cond) {
				break
			}
		}

		result = Eval(node.Body, env)

		if isError(result) {
			return result
		}

		if rv, ok := result.(*object.ReturnValue); ok {
			if str, ok := rv.Value.(*object.String); ok && str.Value == "__BREAK__" {
				break
			}
			return rv
		}

		iterations++
	}

	return result
}

func evalForStatement(node *ast.ForStatement, env *object.Environment) object.Object {
	iterable := Eval(node.Iterable, env)
	if isError(iterable) {
		return iterable
	}

	var result object.Object = NULL

	switch iter := iterable.(type) {
	case *object.Array:
		for _, elem := range iter.Elements {
			env.Set(node.Iterator, elem)
			result = Eval(node.Body, env)

			if isError(result) {
				return result
			}

			if rv, ok := result.(*object.ReturnValue); ok {
				if str, ok := rv.Value.(*object.String); ok && str.Value == "__BREAK__" {
					break
				}
				return rv
			}
		}
	case *object.String:
		for _, ch := range iter.Value {
			env.Set(node.Iterator, &object.String{Value: string(ch)})
			result = Eval(node.Body, env)

			if isError(result) {
				return result
			}

			if rv, ok := result.(*object.ReturnValue); ok {
				if str, ok := rv.Value.(*object.String); ok && str.Value == "__BREAK__" {
					break
				}
				return rv
			}
		}
	default:
		return newError("cannot iterate over %s", iterable.Type())
	}

	return result
}

func evalIfStatement(node *ast.IfStatement, env *object.Environment) object.Object {
	condition := Eval(node.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(node.Consequence, env)
	} else if node.Alternative != nil {
		return Eval(node.Alternative, env)
	}

	return NULL
}

func evalFunctionDefinition(node *ast.FunctionDefinition, env *object.Environment) object.Object {
	fn := &object.Function{
		Parameters: node.Parameters,
		Body:       node.Body,
		Env:        env,
		Name:       node.Name,
	}
	env.Set(node.Name, fn)
	return NULL
}

func evalInfixExpression(node *ast.InfixExpression, env *object.Environment) object.Object {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}

	right := Eval(node.Right, env)
	if isError(right) {
		return right
	}

	switch node.Operator {
	case "+":
		return evalPlusOperatorExpression(left, right)
	case "-":
		return evalMinusOperatorExpression(left, right)
	case "*":
		return evalMultiplyOperatorExpression(left, right)
	case "/":
		return evalDivideOperatorExpression(left, right)
	case "%":
		return evalModuloOperatorExpression(left, right)
	case "==":
		return nativeBoolToBooleanObject(left == right || left.Inspect() == right.Inspect())
	case "!=":
		return nativeBoolToBooleanObject(left != right && left.Inspect() != right.Inspect())
	case "<":
		return evalComparisonExpression(left, right, "<")
	case ">":
		return evalComparisonExpression(left, right, ">")
	case "<=":
		return evalComparisonExpression(left, right, "<=")
	case ">=":
		return evalComparisonExpression(left, right, ">=")
	case "&&", "and":
		return nativeBoolToBooleanObject(isTruthy(left) && isTruthy(right))
	case "||", "or":
		return nativeBoolToBooleanObject(isTruthy(left) || isTruthy(right))
	default:
		return newError("unknown operator: %s %s %s", left.Type(), node.Operator, right.Type())
	}
}

func evalPlusOperatorExpression(left, right object.Object) object.Object {
	switch l := left.(type) {
	case *object.Integer:
		switch r := right.(type) {
		case *object.Integer:
			return &object.Integer{Value: l.Value + r.Value}
		case *object.Float:
			return &object.Float{Value: float64(l.Value) + r.Value}
		}
	case *object.Float:
		switch r := right.(type) {
		case *object.Integer:
			return &object.Float{Value: l.Value + float64(r.Value)}
		case *object.Float:
			return &object.Float{Value: l.Value + r.Value}
		}
	case *object.String:
		switch r := right.(type) {
		case *object.String:
			return &object.String{Value: l.Value + r.Value}
		default:
			return &object.String{Value: l.Value + r.Inspect()}
		}
	}
	return newError("type mismatch: %s + %s", left.Type(), right.Type())
}

func evalMinusOperatorExpression(left, right object.Object) object.Object {
	switch l := left.(type) {
	case *object.Integer:
		switch r := right.(type) {
		case *object.Integer:
			return &object.Integer{Value: l.Value - r.Value}
		case *object.Float:
			return &object.Float{Value: float64(l.Value) - r.Value}
		}
	case *object.Float:
		switch r := right.(type) {
		case *object.Integer:
			return &object.Float{Value: l.Value - float64(r.Value)}
		case *object.Float:
			return &object.Float{Value: l.Value - r.Value}
		}
	}
	return newError("type mismatch: %s - %s", left.Type(), right.Type())
}

func evalMultiplyOperatorExpression(left, right object.Object) object.Object {
	switch l := left.(type) {
	case *object.Integer:
		switch r := right.(type) {
		case *object.Integer:
			return &object.Integer{Value: l.Value * r.Value}
		case *object.Float:
			return &object.Float{Value: float64(l.Value) * r.Value}
		}
	case *object.Float:
		switch r := right.(type) {
		case *object.Integer:
			return &object.Float{Value: l.Value * float64(r.Value)}
		case *object.Float:
			return &object.Float{Value: l.Value * r.Value}
		}
	}
	return newError("type mismatch: %s * %s", left.Type(), right.Type())
}

func evalDivideOperatorExpression(left, right object.Object) object.Object {
	switch l := left.(type) {
	case *object.Integer:
		switch r := right.(type) {
		case *object.Integer:
			if r.Value == 0 {
				return newError("division by zero")
			}
			return &object.Integer{Value: l.Value / r.Value}
		case *object.Float:
			if r.Value == 0 {
				return newError("division by zero")
			}
			return &object.Float{Value: float64(l.Value) / r.Value}
		}
	case *object.Float:
		switch r := right.(type) {
		case *object.Integer:
			if r.Value == 0 {
				return newError("division by zero")
			}
			return &object.Float{Value: l.Value / float64(r.Value)}
		case *object.Float:
			if r.Value == 0 {
				return newError("division by zero")
			}
			return &object.Float{Value: l.Value / r.Value}
		}
	}
	return newError("type mismatch: %s / %s", left.Type(), right.Type())
}

func evalModuloOperatorExpression(left, right object.Object) object.Object {
	switch l := left.(type) {
	case *object.Integer:
		switch r := right.(type) {
		case *object.Integer:
			if r.Value == 0 {
				return newError("division by zero")
			}
			return &object.Integer{Value: l.Value % r.Value}
		}
	}
	return newError("type mismatch: %s %% %s", left.Type(), right.Type())
}

func evalComparisonExpression(left, right object.Object, op string) object.Object {
	var lVal, rVal float64

	switch l := left.(type) {
	case *object.Integer:
		lVal = float64(l.Value)
	case *object.Float:
		lVal = l.Value
	default:
		return newError("cannot compare %s", left.Type())
	}

	switch r := right.(type) {
	case *object.Integer:
		rVal = float64(r.Value)
	case *object.Float:
		rVal = r.Value
	default:
		return newError("cannot compare %s", right.Type())
	}

	var result bool
	switch op {
	case "<":
		result = lVal < rVal
	case ">":
		result = lVal > rVal
	case "<=":
		result = lVal <= rVal
	case ">=":
		result = lVal >= rVal
	}

	return nativeBoolToBooleanObject(result)
}

func evalPrefixExpression(node *ast.PrefixExpression, env *object.Environment) object.Object {
	right := Eval(node.Right, env)
	if isError(right) {
		return right
	}

	switch node.Operator {
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	case "+":
		return right
	case "!":
		return nativeBoolToBooleanObject(!isTruthy(right))
	default:
		return newError("unknown operator: %s%s", node.Operator, right.Type())
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	switch r := right.(type) {
	case *object.Integer:
		return &object.Integer{Value: -r.Value}
	case *object.Float:
		return &object.Float{Value: -r.Value}
	default:
		return newError("unknown operator: -%s", right.Type())
	}
}

func evalCallExpression(node *ast.CallExpression, env *object.Environment) object.Object {
	function := Eval(node.Function, env)
	if isError(function) {
		return function
	}

	args := evalExpressions(node.Arguments, env)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}

	return applyFunction(function, args)
}

func evalMethodCallExpression(node *ast.MethodCallExpression, env *object.Environment) object.Object {
	obj := Eval(node.Object, env)
	if isError(obj) {
		return obj
	}

	args := evalExpressions(node.Arguments, env)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}

	return applyMethod(obj, node.Method, args)
}

func evalIndexExpression(node *ast.IndexExpression, env *object.Environment) object.Object {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}

	index := Eval(node.Index, env)
	if isError(index) {
		return index
	}

	switch l := left.(type) {
	case *object.Array:
		return evalArrayIndexExpression(l, index)
	case *object.String:
		return evalStringIndexExpression(l, index)
	case *object.Hash:
		return evalHashIndexExpression(l, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(array *object.Array, index object.Object) object.Object {
	idx, ok := index.(*object.Integer)
	if !ok {
		return newError("array index must be integer, got %s", index.Type())
	}

	i := idx.Value
	if i < 0 || i >= int64(len(array.Elements)) {
		return NULL
	}

	return array.Elements[i]
}

func evalStringIndexExpression(str *object.String, index object.Object) object.Object {
	idx, ok := index.(*object.Integer)
	if !ok {
		return newError("string index must be integer, got %s", index.Type())
	}

	i := idx.Value
	if i < 0 || i >= int64(len(str.Value)) {
		return NULL
	}

	return &object.String{Value: string(str.Value[i])}
}

func evalHashIndexExpression(hash *object.Hash, index object.Object) object.Object {
	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}

	pair, ok := hash.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}

	return pair.Value
}

func evalPipeExpression(node *ast.PipeExpression, env *object.Environment) object.Object {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}

	switch fn := node.Right.(type) {
	case *ast.Identifier:
		builtin, ok := builtins[fn.Value]
		if ok {
			return builtin.Fn(left)
		}
		return newError("unknown function in pipe: %s", fn.Value)
	case *ast.CallExpression:
		function := Eval(fn.Function, env)
		if isError(function) {
			return function
		}
		args := []object.Object{left}
		otherArgs := evalExpressions(fn.Arguments, env)
		args = append(args, otherArgs...)
		return applyFunction(function, args)
	default:
		return newError("invalid pipe expression")
	}
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: " + node.Value)
}

func evalListLiteral(node *ast.ListLiteral, env *object.Environment) object.Object {
	elements := evalExpressions(node.Elements, env)
	if len(elements) == 1 && isError(elements[0]) {
		return elements[0]
	}
	return &object.Array{Elements: elements}
}

func evalDictLiteral(node *ast.DictLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}

		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		pairs[hashKey.HashKey()] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch function := fn.(type) {
	case *object.Function:
		env := createFunctionEnv(function, args)
		result := Eval(function.Body.(*ast.BlockStatement), env)
		return unwrapReturnValue(result)
	case *object.Builtin:
		return function.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func applyMethod(obj object.Object, method string, args []object.Object) object.Object {
	switch o := obj.(type) {
	case *object.Array:
		return applyArrayMethod(o, method, args)
	case *object.String:
		return applyStringMethod(o, method, args)
	case *object.File:
		return applyFileMethod(o, method, args)
	case *object.Module:
		return applyModuleMethod(o, method, args)
	default:
		return newError("method %s not found on %s", method, obj.Type())
	}
}

func createFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		if paramIdx < len(args) {
			env.Set(param, args[paramIdx])
		} else {
			env.Set(param, NULL)
		}
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	case *object.Integer:
		return obj.Value != 0
	case *object.String:
		return obj.Value != ""
	default:
		return true
	}
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func evalMathExpression(expr string) string {
	expr = strings.TrimSpace(expr)
	
	l := lexer.New(expr)
	p := parser.New(l)
	astNode := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		return expr
	}
	
	env := object.NewEnvironment()
	result := Eval(astNode, env)
	
	if isError(result) {
		return expr
	}
	
	return result.Inspect()
}
