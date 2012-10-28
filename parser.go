package mandira

import (
	"fmt"
	"strconv"
)

/* Parser for the extended features in Mandira.

word = ([a-zA-Z1-9]+)
dot = .
binary = <|<=|>|>=|!=|==|and|or
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
cond = unary varexpr | varexpr
binarycond = cond binary cond [binary cond ...]
condexpr = [(] cond | binarycond [)]
condition = condexpr [binary condexpr...]

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

type varExpr struct {
	exprs []interface{}
}

type parserError struct {
	token   string
	tokens  []string
	message string
}

func (p *parserError) Error() string {
	return fmt.Sprintf("%s: %s in %v", p.message, p.token, p.tokens)
}

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

func parseFuncExpression(tokens []string) (*funcExpr, error) {
	fe := &funcExpr{}
	if len(tokens) == 0 {
		return fe, &parserError{"", tokens, "Expected filter name"}
	}
	fe.name = tokens[0]
	if len(tokens) == 1 {
		return fe, nil
	}
	if tokens[1] != "(" || tokens[len(tokens)-1] != ")" {
		return fe, &parserError{tokens[1], tokens, "Expected ()'s around arguments"}
	}

	for i := 2; i < len(tokens)-1; i++ {
		if i%2 == 0 {
			// TODO: verify valid variable names
			fe.arguments = append(fe.arguments, parseAtom(tokens[i]))
		} else {
			if tokens[i] != "," {
				return fe, &parserError{tokens[i], tokens, "Expected comma (,)"}
			}
		}
	}

	return fe, nil
}

func parseVarExpression(tokens []string) (*varExpr, error) {
	expr := &varExpr{}
	if len(tokens) == 0 {
		return expr, &parserError{"", tokens, "Empty expression"}
	}
	// the first token is definitely a variable
	expr.exprs = append(expr.exprs, &lookupExpr{tokens[0]})
	if len(tokens) == 1 {
		return expr, nil
	}

	if tokens[1] != "|" {
		return expr, &parserError{tokens[1], tokens, "Expected pipe (|) for filter expression"}
	}

	i := 2
	run := 2
	for ; i < len(tokens); i++ {
		if tokens[i] == "|" && run == i {
			return expr, &parserError{tokens[i], tokens, "Expected filter expression"}
		} else if tokens[i] == "|" {
			e, err := parseFuncExpression(tokens[run:i])
			if err != nil {
				return expr, err
			}
			expr.exprs = append(expr.exprs, e)
			run = i + 1
		}
	}
	if run != i {
		e, err := parseFuncExpression(tokens[run:])
		if err != nil {
			return expr, err
		}
		expr.exprs = append(expr.exprs, e)
	}
	return expr, nil
}

// tokenize a conditional, return a list of strings
func tokenize(c string) ([]string, error) {
	b := []byte(c)
	tn := struct {
		tokens []string
		run    int
		p      int
	}{
		[]string{},
		0,
		0,
	}

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
