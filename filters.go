package mandira

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

var filters = map[string]interface{}{}

func AddFilter(filter interface{}, name ...string) {
	/* FIXME: typecheck */
	var fname string
	if len(name) == 1 {
		fname = name[0]
	} else {
		val := reflect.ValueOf(filter)
		// this returns something like: /jmoiron/devel/mandira.Len
		spl := strings.Split(runtime.FuncForPC(val.Pointer()).Name(), ".")
		fname = strings.ToLower(spl[len(spl)-1])
	}

	filters[fname] = filter
}

// Return a filter (or nil)
func GetFilter(name string) interface{} {
	filter, ok := filters[name]
	if !ok {
		return nil
	}
	return filter
}

// Return the length of the argument, or 0 if that is not a valid action
func Len(arg interface{}) int {
	val := reflect.ValueOf(arg)
	switch val.Kind() {
	case reflect.Array, reflect.Slice, reflect.String, reflect.Map:
		return val.Len()
	}
	return 0
}

// Return the index of the argument at arg I
func Index(arg interface{}, idx_ interface{}) interface{} {
	idx := int(idx_.(int64))
	val := reflect.ValueOf(arg)
	switch val.Kind() {
	case reflect.Array, reflect.Slice:
		if val.Len() > idx {
			return val.Index(idx).Interface()
		}
	case reflect.String:
		s := val.String()
		if len(s) > idx {
			return string(s[idx])
		}
	}
	return ""
}

func Format(arg interface{}, format string) interface{} {
	return fmt.Sprintf(format, arg)
}

func Date(arg interface{}, format string) interface{} {
	return ""
}

func Join(arg interface{}, joiner string) string {
	slist := []string{}
	val := reflect.ValueOf(arg)
	switch val.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			slist = append(slist, fmt.Sprint(val.Index(i).Interface()))
		}
	}
	return strings.Join(slist, joiner)
}

func DivisibleBy(base_, by_ interface{}) bool {
	base := reflect.ValueOf(base_).Int()
	by := reflect.ValueOf(by_).Int()
	return base%by == 0
}

func init() {
	AddFilter(strings.ToUpper, "upper")
	AddFilter(strings.ToLower, "lower")
	AddFilter(strings.Title, "title")
	AddFilter(Len)
	AddFilter(Index)
	AddFilter(Format)
	AddFilter(Date)
	AddFilter(Join)
	AddFilter(DivisibleBy)
}
