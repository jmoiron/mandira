// Evaluation helpers for Mandira

package mandira

import (
	"errors"
	"fmt"
	"reflect"
)

/*
type lookupExpr struct {
	name string
}

type funcExpr struct {
	name string
	// these are either literal values or lookup expressions
	arguments []interface{}
}

type cond struct {
	not  bool
	expr interface{}
}

type bincond struct {
	oper string
	lhs  *cond
	rhs  *cond
}

type condExpr struct {
	oper string
	lhs  interface{}
	rhs  interface{}
}

type varExpr struct {
	exprs []interface{}
}
*/

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

// Evaluate a unary condition;  evaluates either to the value of the expression
// or a boolean (tested with isNil) if the expression is a negation
func (c *cond) Eval(contexts []interface{}) (interface{}, error) {
	var exprval interface{}
	switch c.expr.(type) {
	case *varExpr:
		exprval, _ = c.expr.(*varExpr).Eval(contexts)
	case string:
		exprval = fmt.Sprint(c.expr)
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

// Evaluate a binary condition;  evaluates to a boolean value
func (c *bincond) Eval(contexts []interface{}) bool {
	lhs, _ := c.lhs.Eval(contexts)
	rhs, _ := c.rhs.Eval(contexts)

	vl := reflect.ValueOf(lhs)
	vr := reflect.ValueOf(rhs)

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic while doing comparisson: %v, %v", lhs, rhs)
		}
	}()

	switch vl.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return compInt(c.oper, vl.Int(), vr.Int())
	case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return compInt(c.oper, int64(vl.Uint()), int64(vr.Uint()))
	case reflect.String:
		return compString(c.oper, vl.String(), vr.String())
	}

	return false
}

func (c *condExpr) Eval(contexts []interface{}) bool {
	var lhsval interface{}

	switch c.lhs.(type) {
	case *bincond:
		lhsval = c.lhs.(*bincond).Eval(contexts)
	case *cond:
		lhsval, _ = c.lhs.(*cond).Eval(contexts)
	default:
		return false
	}

	if len(c.oper) == 0 {
		return !isNil(reflect.ValueOf(lhsval))
	}

	var rhsval interface{}
	switch c.rhs.(type) {
	case *bincond:
		rhsval = c.rhs.(*bincond).Eval(contexts)
	case *cond:
		rhsval, _ = c.rhs.(*cond).Eval(contexts)
	default:
		return false
	}

	switch c.oper {
	case "and":
		return !isNil(reflect.ValueOf(lhsval)) && !isNil(reflect.ValueOf(rhsval))
	case "or":
		return !isNil(reflect.ValueOf(lhsval)) || !isNil(reflect.ValueOf(rhsval))
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
			/* FIXME: check non-string types for context args in filters */
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
