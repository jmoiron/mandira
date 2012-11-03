package mandira

import (
	"fmt"
	"strconv"
)

/* Parser for the extended features in Mandira.

word = ([a-zA-Z1-9]+)
binop = <|<=|>|>=|!=|==
comb = or|and
unary = not
filter = |
variable = word
string = " .* "
atom = variable | string | word
funcexpr = word [( atom[, atom...] )]
varexpr = variable [|funcexpr...]

Conditional logic is mostly as expected, with operators of the same precedence
being computed from left to right.  

They are, from low to high: binops, combs, unary, parens

In the future, "and" may be higher priority than "or".

*/

// A lookup expression is a naked word which will be looked up in the context at render time
type lookupExpr struct {
	name string
}

// A varExpr is a lookupExpr followed by zero or more funcExprs
type varExpr struct {
	exprs []interface{}
}

// A func expression has a function name to be looked up in the filter list 
// at render time and a list of arguments, which are varExprs or literals
type funcExpr struct {
	name      string
	arguments []interface{}
}

// A cond is a unary condition with a single value and optional negation
type cond struct {
	not  bool
	expr interface{}
}

// A conditional is a n-ary conditional with n opers and n+1 expressions, which
// can be conds or conditionals
type conditional struct {
	not   bool
	opers []string
	exprs []interface{}
}

// A list of tokens with a pointer (p) and a run (run)
// In tokenizing, this structure tracks tokens, but p points to the []byte
// being tokenized, and run keeps track of the length of the current token
// In parsing, this p is used as a pointer to a token in tokens
type tokenList struct {
	tokens []string
	p      int
	run    int
}

// Return the number of remaining tokens
func (t *tokenList) Remaining() int {
	return len(t.tokens) - t.p
}

// Return the next token.  Returns "" if there are none left.
func (t *tokenList) Next() string {
	if t.p == len(t.tokens) {
		return ""
	}
	t.p++
	return t.tokens[t.p-1]
}

// Peek at the current token. Returns "" if there are none left.
func (t *tokenList) Peek() string {
	if t.p == len(t.tokens) {
		return ""
	}
	return t.tokens[t.p]
}

// Go back to previous token and return it.
func (t *tokenList) Prev() string {
	if t.p > 0 {
		t.p--
	}
	return t.tokens[t.p]
}

type parserError struct {
	tokens  *tokenList
	message string
}

func (p *parserError) Error() string {
	return fmt.Sprintf(`%s: "%s" in %v`, p.message, p.tokens.Peek(), p.tokens)
}

// Parse an atom;  an atom is a literal or a lookup expression.
func parseAtom(token string) interface{} {
	if token[0] == '"' {
		return token[1 : len(token)-1]
	}
	i, err := strconv.ParseInt(token, 10, 64)
	if err == nil {
		return i
	}
	f, err := strconv.ParseFloat(token, 64)
	if err == nil {
		return f
	}
	return &lookupExpr{token}
}

// parse a value, which is a literal or a variable expression
func parseValue(tokens *tokenList) (interface{}, error) {
	tok := tokens.Next()
	if len(tok) == 0 {
		return nil, &parserError{tokens, "Expected a value, found nothing"}
	}
	try := parseAtom(tok)
	/* if this wasn't a lookupExpr, then it's a literal */
	if _, ok := try.(*lookupExpr); !ok {
		return try, nil
	}
	tokens.Prev()
	varexp, err := parseVarExpression(tokens)
	return varexp, err
}

// parse a value and return a unary cond expr (to negate values)
func parseCond(tokens *tokenList) (*cond, error) {
	var err error
	c := &cond{}
	c.expr, err = parseValue(tokens)
	return c, err
}

// Parse a conditional expression, recurse each time a paren is encountered
func parseCondition(tokens *tokenList) (*conditional, error) {
	c := &conditional{}
	negated := false
	expectCond := true

	for tok := tokens.Next(); len(tok) > 0; tok = tokens.Next() {
		switch tok {
		case "(":
			if !expectCond {
				return c, &parserError{tokens, "Expected an operator, not a " + tok}
			}
			expr, err := parseCondition(tokens)
			if err != nil {
				return c, err
			}
			expr.not = negated
			c.exprs = append(c.exprs, expr)

			negated = false
			expectCond = false
		case "not":
			if !expectCond {
				return c, &parserError{tokens, "Expected an operator, not a " + tok}
			}
			negated = !negated
		case ")":
			if expectCond {
				return c, &parserError{tokens, "Expected a condition, not a " + tok}
			}
			return c, nil
		case "or", "and", ">", "<", "<=", ">=", "==", "!=":
			if expectCond {
				return c, &parserError{tokens, "Expected a condition, not an operator " + tok}
			}
			c.opers = append(c.opers, tok)
			expectCond = true
		default:
			if !expectCond {
				return c, &parserError{tokens, "Expected an operator, not " + tok}
			}
			tokens.Prev()
			expr, err := parseCond(tokens)
			if err != nil {
				return c, err
			}
			expr.not = negated
			c.exprs = append(c.exprs, expr)
			// reset everything
			expectCond = false
			negated = false
		}
	}

	return c, nil
}

// parse a function expression, which comes after each | in a filter
func parseFuncExpression(tokens *tokenList) (*funcExpr, error) {
	fe := &funcExpr{}
	fe.name = tokens.Next()
	if len(fe.name) == 0 {
		return fe, &parserError{tokens, "Expected filter name, got nil"}
	}
	tok := tokens.Peek()
	if tok == "(" {
		tokens.Next()
		for tok = tokens.Next(); len(tok) > 0; tok = tokens.Next() {
			fe.arguments = append(fe.arguments, parseAtom(tok))
			tok = tokens.Next()
			if tok == ")" {
				break
			}
			if tok != "," {
				return fe, &parserError{tokens, "Expected comma (,)"}
			}
		}
	}

	return fe, nil
}

// parse a variable expression, which is a lookup + 0 or more func exprs
func parseVarExpression(tokens *tokenList) (*varExpr, error) {

	expr := &varExpr{}
	tok := tokens.Next()
	if len(tok) == 0 {
		return expr, &parserError{tokens, "Empty expression"}
	}
	// the first token is definitely a variable
	expr.exprs = append(expr.exprs, &lookupExpr{tok})

	tok = tokens.Next()
	if tok != "|" && tok != "" {
		tokens.Prev()
		return expr, nil
	}

	for len(tok) > 0 {
		if tok == "|" {
			e, err := parseFuncExpression(tokens)
			if err != nil {
				return expr, err
			}
			expr.exprs = append(expr.exprs, e)
			tok = tokens.Next()
		} else if tok == "" {
			return expr, nil
		} else {
			tokens.Prev()
			return expr, nil
		}
	}

	return expr, nil
}

// Parse aa "variable element", which returns a varElement (AST)
func parseVarElement(s string) (*varElement, error) {
	var elem = &varElement{}
	tokens, err := tokenize(s)
	if err != nil {
		return elem, err
	}
	expr, err := parseVarExpression(&tokenList{tokens, 0, 0})
	if err != nil {
		return elem, err
	}
	elem.expr = expr
	return elem, nil
}

// Parse a "conditional element", which returns a conditional section element (AST)
func parseCondElement(s string) (*sectionElement, error) {
	var elem = &sectionElement{}
	tokens, err := tokenize(s)
	if err != nil {
		return elem, err
	}
	expr, err := parseCondition(&tokenList{tokens, 0, 0})
	if err != nil {
		return elem, err
	}
	elem.expr = expr
	return elem, nil
}

// tokenize an expression, returning a list of strings or an error
func tokenize(c string) ([]string, error) {
	b := []byte(c)
	tn := tokenList{[]string{}, 0, 0}

	for ; tn.p < len(b); tn.p++ {
		switch b[tn.p] {
		case ' ', '\t':
			if tn.run < tn.p {
				tn.tokens = append(tn.tokens, string(b[tn.run:tn.p]))
			}
			tn.run = tn.p + 1
		/* tokens which can be singular or double */
		case '<', '>':
			if tn.run < tn.p {
				tn.tokens = append(tn.tokens, string(b[tn.run:tn.p]))
			}
			if tn.p+1 < len(b) && b[tn.p+1] == '=' {
				tn.tokens = append(tn.tokens, string(b[tn.p:tn.p+2]))
				tn.p++
			} else {
				tn.tokens = append(tn.tokens, string(b[tn.p]))
			}
			tn.run = tn.p + 1
		/* tokens which must be double */
		case '!', '=':
			if tn.run < tn.p {
				tn.tokens = append(tn.tokens, string(b[tn.run:tn.p]))
			}
			if tn.p+1 < len(b) && b[tn.p+1] == '=' {
				tn.tokens = append(tn.tokens, string(b[tn.p:tn.p+2]))
				tn.p++
			} else {
				return tn.tokens, parseError{tn.p, "invalid token: " + string(b[tn.p])}
			}
			tn.run = tn.p + 1
		case '"':
			start := tn.p
			tn.p++
			for ; tn.p < len(b); tn.p++ {
				if b[tn.p] == '"' && b[tn.p-1] != '\\' {
					tn.tokens = append(tn.tokens, string(b[start:tn.p+1]))
					break
				}
			}
			tn.run = tn.p + 1
		/* tokens which are only ever single */
		case '|', '(', ')', ',':
			if tn.p > 0 && b[tn.p] == '\\' && b[tn.p] == '"' {
				tn.run = tn.p + 1
				continue
			}
			if tn.run < tn.p {
				tn.tokens = append(tn.tokens, string(b[tn.run:tn.p]))
			}
			tn.tokens = append(tn.tokens, string(b[tn.p]))
			tn.run = tn.p + 1
		default:

		}
	}

	if tn.run < len(b) {
		tn.tokens = append(tn.tokens, string(b[tn.run:]))
	}
	return tn.tokens, nil
}
