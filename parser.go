package mandira

import (
	"fmt"
	"strconv"
)

/* Parser for the extended features in Mandira.

word = ([a-zA-Z1-9]+)
dot = .
binop = <|<=|>|>=|!=|==
comb = or|and
unary = not
filter = |
variable = word
string = " .* "
int = [0-9]+
float = int dot [int]
bool = true | false
atom = variable | string | int | float | bool
funcexpr = word [( atom[, atom...] )]
varexpr = variable [|funcexpr...]

The following grammar describes a simplified, easy to implement boolean algebra.

Conditions can be a varexpr (var|filter...) or a negated var expr.
Binary Conditions must be two of these joined by a numeric binop
ConditionalExpressions are either one Condition (either type), or two joined by
a boolean combinator (and/or).

Multiple and/or is prohibited.  Grouping is unnecessary.  To achieve more complex
logic, use multiple conditional blocks. Numeric operators take precedence over and/or.
Binary logic is evaluated from left to right, and & ors are short circuited by
false & true values in the lhs, respectively.

cond = varexpr | not varexpr
bincond = cond binop cond
boolexpr = cond | bincond
condexpr = boolexpr [comb boolexpr]

*/

// -- Literals 

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
// lookup expressions can have any text but whitespace in them
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

// parse a single unary condition (or naked varexpr)
func parseCond(tokens *tokenList) (*cond, error) {
	var err error
	c := &cond{}
	tok := tokens.Peek()
	if tok == "not" {
		c.not = true
		tokens.Next()
	}
	c.expr, err = parseVarExpression(tokens)
	return c, err
}

// parse a unary or binary boolean expression and return it.
// Valid return values are of type cond or bincond
func parseBoolExpression(tokens *tokenList) (interface{}, error) {
	c, err := parseCond(tokens)
	if err != nil {
		return c, err
	}
	tok := tokens.Peek()
	switch tok {
	case "<", "<=", ">", ">=", "==", "!=":
		tokens.Next()
		c2, err := parseCond(tokens)
		if err != nil {
			return c2, err
		}
		// disallow things like "not foo > bar" in favor of "foo <= bar"
		// and also nonsensical stuff like "not foo > not bar"
		if c.not || c2.not {
			return nil, &parserError{tokens, "Unary operators invalid in binary conditions (use converse of binary operator instead)"}
		}
		return &bincond{tok, c, c2}, nil
	}
	return c, nil
}

// Parse a full condition expression:
func parseCondExpression(tokens *tokenList) (*condExpr, error) {
	var err error
	c := &condExpr{}
	c.lhs, err = parseBoolExpression(tokens)

	if err != nil {
		return &condExpr{}, err
	}

	tok := tokens.Peek()
	switch tok {
	case "and", "or":
		c.oper = tok
		tokens.Next()
		c.rhs, err = parseBoolExpression(tokens)
		if err != nil {
			return c, err
		}
		return c, nil
	}
	return c, nil
}

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

// tokenize a conditional, return a list of strings
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
