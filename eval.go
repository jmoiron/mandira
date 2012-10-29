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
			fmt.Println("Unknown arg type")
		}
	}

	retval := filterVal.Call(argvals)[0]
	return retval.Interface(), nil
}

// Evaluate a varExpr given the contexts.  Return a string and possible error
func (v *varExpr) Eval(contexts []interface{}) (string, error) {
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
	return fmt.Sprint(inter), nil
}
