package mandira

import (
	"testing"
)

func TestTokenizer(t *testing.T) {
	type MS map[string][]string
	tests := []MS{
		MS{"bare": []string{"bare"}},
		MS{" bare": []string{"bare"}},
		MS{"bare ": []string{"bare"}},
		MS{" bare ": []string{"bare"}},
		MS{">bare": []string{">", "bare"}},
		MS{"<bare": []string{"<", "bare"}},
		MS{">=bare": []string{">=", "bare"}},
		MS{"<=bare": []string{"<=", "bare"}},
		MS{"!=bare": []string{"!=", "bare"}},
		MS{"==bare": []string{"==", "bare"}},
		MS{"bare >= bare": []string{"bare", ">=", "bare"}},
		MS{"bare>=bare": []string{"bare", ">=", "bare"}},
		MS{"bare>= bare": []string{"bare", ">=", "bare"}},
		MS{"bare >=bare": []string{"bare", ">=", "bare"}},
		MS{"i>a": []string{"i", ">", "a"}},
		MS{`bare|func("foo", bar, 1.5) >= 9`: []string{"bare", "|", "func", "(", `"foo"`, ",", "bar", ",", "1.5", ")", ">=", "9"}},
		MS{`b|func("foo bar, 今日は世界")`: []string{"b", "|", "func", "(", `"foo bar, 今日は世界"`, ")"}},
	}
	errs := []string{
		"a = b", // single = is an invalid token
		"!a",    // single ! is an invalid token
	}

	for _, test := range tests {
		for k, v := range test {
			tokens, err := tokenize(k)
			if err != nil {
				t.Errorf("Got an error tokenizing: %v\n", err)
				continue
			}
			if len(tokens) != len(v) {
				t.Errorf("Wrong number of tokens: %v vs %v in \"%v\"\n", tokens, v, k)
				continue
			}
			for i, tok := range tokens {
				if tok != v[i] {
					t.Errorf("Expected %v, got %v in \"%v\"\n", v[i], tok, k)
				}
			}
		}
	}
	for _, e := range errs {
		_, err := tokenize(e)
		if err == nil {
			t.Errorf("Expected tokenizer error on \"%v\"\n", e)
		}
	}
}

func tErr(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Unexpected error: %v\n", err)
	}
}

func TestFuncParser(t *testing.T) {

}

func TestVarParser(t *testing.T) {
	// naked test
	expr, err := parseVarExpression([]string{"hello"})
	tErr(t, err)
	if len(expr.exprs) != 1 {
		t.Errorf("Expected a single lookup expression, got %v\n", expr.exprs)
	}
	lu, ok := expr.exprs[0].(*lookupExpr)
	if !ok {
		t.Errorf("Expected a lookup expression, got %v\n", expr.exprs)
	}
	if lu.name != "hello" {
		t.Errorf("Expected varname to be \"hello\", got %s\n", lu.name)
	}

	// naked with filter
	expr, err = parseVarExpression([]string{"hello", "|", "upper"})
	tErr(t, err)
	if len(expr.exprs) != 2 {
		t.Fatalf("Expected 2 expressions, got %v\n", len(expr.exprs))
	}
	lu, ok = expr.exprs[0].(*lookupExpr)
	if !ok {
		t.Errorf("Expected a lookup expression, got %v\n", expr.exprs)
	}
	if lu.name != "hello" {
		t.Errorf("Expected varname to be \"hello\", got %s\n", lu.name)
	}
	fu, ok := expr.exprs[1].(*funcExpr)
	if !ok {
		t.Errorf("Expected a func expression, got %v\n", expr.exprs)
	}
	if fu.name != "upper" {
		t.Errorf("Expected funcname to be \"upper\", got %s\n", fu.name)
	}
	if len(fu.arguments) != 0 {
		t.Errorf("Got unexpected arguments (%v)\n", fu.arguments)
	}

	// filter chain with arguments
	toks, _ := tokenize(`hello|upper|join(", ", 3.5, someVar)|fake("hi")`)
	if len(toks) != 17 {
		t.Fatalf("Unexpected tokenization results for test %v\n", toks)
	}
	expr, err = parseVarExpression(toks)

	if len(expr.exprs) != 4 {
		t.Fatalf("Expected 4 expressions, got %d (%v)\n", len(expr.exprs), expr.exprs)
	}
	lu, ok = expr.exprs[0].(*lookupExpr)
	if !ok {
		t.Errorf("Expected a lookup expression, got %v\n", expr.exprs)
	}
	if lu.name != "hello" {
		t.Errorf("Expected varname to be \"hello\", got %s\n", lu.name)
	}
	fu, ok = expr.exprs[1].(*funcExpr)
	if !ok {
		t.Errorf("Expected a func expression, got %v\n", expr.exprs)
	}
	if fu.name != "upper" {
		t.Errorf("Expected funcname to be \"upper\", got %s\n", fu.name)
	}
	if len(fu.arguments) != 0 {
		t.Errorf("Got unexpected arguments (%v)\n", fu.arguments)
	}
	fu, ok = expr.exprs[2].(*funcExpr)
	if !ok {
		t.Errorf("Expected a func expression, got %v\n", expr.exprs)
	}
	if fu.name != "join" {
		t.Errorf("Expected funcname to be \"join\", got %s\n", fu.name)
	}
	if len(fu.arguments) != 3 {
		t.Errorf("Got unexpected number of arguments, expected 3 (%v)\n", fu.arguments)
	}
	s, ok := fu.arguments[0].(string)
	if !ok {
		t.Errorf("Expecting string as first arg\n")
	}
	if s != ", " {
		t.Errorf(`Expecting ", ", got %s`+"\n", s)
	}
	f, ok := fu.arguments[1].(float64)
	if !ok {
		t.Errorf("Expecting float as second arg\n")
	}
	if f != 3.5 {
		t.Errorf("Expecting 3.5, got %v\n", f)
	}
	lu, ok = fu.arguments[2].(*lookupExpr)
	if !ok {
		t.Errorf("Expecting lookupExpr as third arg\n")
	}
	if lu.name != "someVar" {
		t.Errorf("Expecting name \"someVar\", got %s\n", lu.name)
	}
	fu, ok = expr.exprs[3].(*funcExpr)
	if !ok {
		t.Errorf("Expected a func expression, got %v\n", expr.exprs)
	}
	if fu.name != "fake" {
		t.Errorf("Expected funcname to be \"fake\", got %s\n", fu.name)
	}
	if len(fu.arguments) != 1 {
		t.Errorf("Got unexpected number of arguments, expected 3 (%v)\n", fu.arguments)
	}
	s, ok = fu.arguments[0].(string)
	if !ok {
		t.Errorf("Expecting string as first arg\n")
	}
	if s != "hi" {
		t.Errorf(`Expecting "hi", got %s`+"\n", s)
	}

}
