// Evaluation helpers for Mandira

package mandira

import (
	"errors"
	"fmt"
	"reflect"
)

func compInt(oper string, l, r int64) bool {
	switch oper {
	case ">":
		return l > r
	case ">=":
		return l >= r
	case "<":
		return l < r
	case "<=":
		return l <= r
	case "!=":
		return l != r
	case "==":
		return l == r
	}
	return false
}

func compString(oper, l, r string) bool {
	switch oper {
	case ">":
		return l > r
	case ">=":
		return l >= r
	case "<":
		return l < r
	case "<=":
		return l <= r
	case "!=":
		return l != r
	case "==":
		return l == r
	}
	return false

}

// run eval for something which is a cond or a conditional
func Eval(expr interface{}, contexts []interface{}) (interface{}, error) {
	switch expr.(type) {
	case *cond:
		return expr.(*cond).Eval(contexts)
	case *conditional:
		return expr.(*conditional).Eval(contexts), nil
	case bool:
		return expr.(bool), nil
	default:
		fmt.Printf("Got unknown type for interface %v: %s", expr, reflect.ValueOf(expr).Kind())
	}
	return nil, nil
}

// Evaluate a unary condition;  evaluates either to the value of the expression
// or a boolean (tested with isNil) if the expression is a negation
func (c *cond) Eval(contexts []interface{}) (interface{}, error) {
	var exprval interface{}
	switch c.expr.(type) {
	case *varExpr:
		exprval, _ = c.expr.(*varExpr).Eval(contexts)
	case string:
		exprval = c.expr.(string)
	case int64:
		exprval = c.expr.(int64)
	case float64:
		exprval = c.expr.(float64)
	}

	if c.not {
		return isNil(reflect.ValueOf(exprval)), nil
	}
	return exprval, nil
}

func (c *conditional) Eval(contexts []interface{}) bool {
	// fast path for single expression conditional exprs like (foo)
	if len(c.opers) == 0 {
		val, err := Eval(c.exprs[0], contexts)
		if err != nil {
			return false
		}
		ret := !isNil(reflect.ValueOf(val))
		if c.not {
			return !ret
		}
		return ret
	}

	// apply numeric logic operators first

	var lhs interface{}
	var rhs interface{}

	var opers []string
	var exprs []interface{}
	reduced := false

	for i, oper := range c.opers {
		if reduced {
			reduced = false
			opers = append(opers, oper)
			continue
		}
		lhs = c.exprs[i]
		rhs = c.exprs[i+1]
		switch oper {
		case "or", "and":
			opers = append(opers, oper)
			exprs = append(exprs, lhs)
		default:
			value := CompEval(oper, lhs, rhs, contexts)
			exprs = append(exprs, value)
			reduced = true
		}
	}
	if !reduced {
		exprs = append(exprs, c.exprs[len(c.exprs)-1])
	}

	if len(opers) == 0 {
		ret := !isNil(reflect.ValueOf(exprs[0]))
		if c.not {
			return !ret
		}
		return ret

	}

	lhs = exprs[0]
	for i, oper := range opers {
		rhs = exprs[i+1]
		lhs = BoolEval(oper, lhs, rhs, contexts)
	}

	if c.not {
		return !lhs.(bool)
	}
	return lhs.(bool)
}

func BoolEval(oper string, lhs, rhs interface{}, contexts []interface{}) bool {
	lhsv, err := Eval(lhs, contexts)
	if err != nil {
		fmt.Printf("Error: %q\n", err)
		lhsv = false
	}
	rhsv, err := Eval(rhs, contexts)
	if err != nil {
		fmt.Printf("Error: %q\n", err)
		rhsv = false
	}

	switch oper {
	case "and":
		return !isNil(reflect.ValueOf(lhsv)) && !isNil(reflect.ValueOf(rhsv))
	case "or":
		return !isNil(reflect.ValueOf(lhsv)) || !isNil(reflect.ValueOf(rhsv))
	}
	return false
}

func CompEval(oper string, lhs, rhs interface{}, contexts []interface{}) bool {
	lhsv, err := Eval(lhs, contexts)
	if err != nil {
		fmt.Printf("Error: %q\n", err)
		return false
	}
	rhsv, err := Eval(rhs, contexts)
	if err != nil {
		fmt.Printf("Error: %q\n", err)
		return false
	}

	vl := reflect.ValueOf(lhsv)
	vr := reflect.ValueOf(rhsv)

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic while doing comparisson: %v, %v", lhs, rhs)
		}
	}()

	switch vl.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return compInt(oper, vl.Int(), vr.Int())
	case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return compInt(oper, int64(vl.Uint()), int64(vr.Uint()))
	case reflect.String:
		return compString(oper, vl.String(), vr.String())
	}

	return false
}

// Apply a filter to a value
func (f *funcExpr) Apply(contexts []interface{}, input interface{}) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic while applying filter %q: %v, %v", f.name, input, f.arguments)
		}
	}()

	filter := GetFilter(f.name)
	if filter == nil {
		return nil, fmt.Errorf("Could not find filter: %s", f.name)
	}

	filterVal := reflect.ValueOf(filter)
	filterType := filterVal.Type()

	argvals := []reflect.Value{reflect.ValueOf(input)}
	for i, arg := range f.arguments {
		switch arg.(type) {
		case string, int64, int, float64:
			argvals = append(argvals, reflect.ValueOf(arg))
		case *lookupExpr:
			lu := arg.(*lookupExpr)
			val := lookup(contexts, lu.name)
			if !val.IsValid() {
				return "", fmt.Errorf("Invalid lookup for filter argument: %s", lu.name)
			}
			argtype := filterType.In(i + 1)
			switch argtype.Kind() {
			case reflect.String:
				argvals = append(argvals, reflect.ValueOf(fmt.Sprint(val.Interface())))
			case reflect.Int, reflect.Int64:
				argvals = append(argvals, reflect.ValueOf(val.Int()))
			}
		default:
			fmt.Printf("Unknown arg type %v\n", arg)
		}
	}

	retval := filterVal.Call(argvals)[0]
	return retval.Interface(), nil
}

// Evaluate a varExpr given the contexts.  Return a string and possible error
func (v *varExpr) Eval(contexts []interface{}) (interface{}, error) {
	var err error
	expr := v.exprs[0].(*lookupExpr)
	val := lookup(contexts, expr.name)

	if !val.IsValid() {
		return "", errors.New("Invalid value in lookup.")
	}

	inter := val.Interface()

	for _, exp := range v.exprs[1:] {
		filter := exp.(*funcExpr)
		inter, err = filter.Apply(contexts, inter)
		if err != nil {
			return "", err
		}
	}
	if inter == nil {
		return "", nil
	}
	return inter, nil
}
