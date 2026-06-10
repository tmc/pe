package promptfoo

import (
	"fmt"
	"strings"
	"unicode"
)

// ComputeDerivedMetrics evaluates each DerivedMetric's expression over the
// given named scores and returns the resulting metric values. The special
// identifier __count resolves to count, the number of contributing results.
//
// Expressions support +, -, *, /, parentheses, numeric literals, and named
// score identifiers. This covers promptfoo's common derivedMetrics cases
// (e.g. "Consistency * 2", "TotalScore / __count") without a third-party
// expression engine. An expression referencing an unknown identifier, or one
// that fails to parse, is skipped and reported in the returned error so the
// caller can surface it without aborting the run.
func ComputeDerivedMetrics(metrics []DerivedMetric, namedScores map[string]float64, count int) (map[string]float64, error) {
	if len(metrics) == 0 {
		return nil, nil
	}
	env := make(map[string]float64, len(namedScores)+1)
	for k, v := range namedScores {
		env[k] = v
	}
	env["__count"] = float64(count)

	out := make(map[string]float64, len(metrics))
	var errs []string
	for _, m := range metrics {
		v, err := evalExpr(m.Value, env)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", m.Name, err))
			continue
		}
		out[m.Name] = v
		// A derived metric is itself available to later expressions.
		env[m.Name] = v
	}
	if len(errs) > 0 {
		return out, fmt.Errorf("derived metrics: %s", strings.Join(errs, "; "))
	}
	return out, nil
}

// exprParser is a minimal recursive-descent evaluator for arithmetic over a
// named-score environment. Grammar:
//
//	expr   := term (('+' | '-') term)*
//	term   := factor (('*' | '/') factor)*
//	factor := number | identifier | '(' expr ')' | '-' factor
type exprParser struct {
	input string
	pos   int
	env   map[string]float64
}

func evalExpr(s string, env map[string]float64) (float64, error) {
	p := &exprParser{input: s, env: env}
	p.skipSpace()
	if p.pos >= len(p.input) {
		return 0, fmt.Errorf("empty expression")
	}
	v, err := p.parseExpr()
	if err != nil {
		return 0, err
	}
	p.skipSpace()
	if p.pos != len(p.input) {
		return 0, fmt.Errorf("unexpected token at %q", p.input[p.pos:])
	}
	return v, nil
}

func (p *exprParser) parseExpr() (float64, error) {
	v, err := p.parseTerm()
	if err != nil {
		return 0, err
	}
	for {
		p.skipSpace()
		if p.pos >= len(p.input) {
			return v, nil
		}
		switch p.input[p.pos] {
		case '+':
			p.pos++
			r, err := p.parseTerm()
			if err != nil {
				return 0, err
			}
			v += r
		case '-':
			p.pos++
			r, err := p.parseTerm()
			if err != nil {
				return 0, err
			}
			v -= r
		default:
			return v, nil
		}
	}
}

func (p *exprParser) parseTerm() (float64, error) {
	v, err := p.parseFactor()
	if err != nil {
		return 0, err
	}
	for {
		p.skipSpace()
		if p.pos >= len(p.input) {
			return v, nil
		}
		switch p.input[p.pos] {
		case '*':
			p.pos++
			r, err := p.parseFactor()
			if err != nil {
				return 0, err
			}
			v *= r
		case '/':
			p.pos++
			r, err := p.parseFactor()
			if err != nil {
				return 0, err
			}
			if r == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			v /= r
		default:
			return v, nil
		}
	}
}

func (p *exprParser) parseFactor() (float64, error) {
	p.skipSpace()
	if p.pos >= len(p.input) {
		return 0, fmt.Errorf("unexpected end of expression")
	}
	c := p.input[p.pos]
	switch {
	case c == '(':
		p.pos++
		v, err := p.parseExpr()
		if err != nil {
			return 0, err
		}
		p.skipSpace()
		if p.pos >= len(p.input) || p.input[p.pos] != ')' {
			return 0, fmt.Errorf("missing closing paren")
		}
		p.pos++
		return v, nil
	case c == '-':
		p.pos++
		v, err := p.parseFactor()
		if err != nil {
			return 0, err
		}
		return -v, nil
	case c >= '0' && c <= '9' || c == '.':
		return p.parseNumber()
	case isIdentStart(rune(c)):
		return p.parseIdent()
	default:
		return 0, fmt.Errorf("unexpected character %q", string(c))
	}
}

func (p *exprParser) parseNumber() (float64, error) {
	start := p.pos
	for p.pos < len(p.input) {
		c := p.input[p.pos]
		if (c >= '0' && c <= '9') || c == '.' {
			p.pos++
			continue
		}
		break
	}
	var v float64
	if _, err := fmt.Sscanf(p.input[start:p.pos], "%g", &v); err != nil {
		return 0, fmt.Errorf("invalid number %q", p.input[start:p.pos])
	}
	return v, nil
}

func (p *exprParser) parseIdent() (float64, error) {
	start := p.pos
	for p.pos < len(p.input) {
		c := rune(p.input[p.pos])
		if isIdentStart(c) || unicode.IsDigit(c) {
			p.pos++
			continue
		}
		break
	}
	name := p.input[start:p.pos]
	v, ok := p.env[name]
	if !ok {
		return 0, fmt.Errorf("unknown identifier %q", name)
	}
	return v, nil
}

func (p *exprParser) skipSpace() {
	for p.pos < len(p.input) && unicode.IsSpace(rune(p.input[p.pos])) {
		p.pos++
	}
}

func isIdentStart(c rune) bool {
	return c == '_' || unicode.IsLetter(c)
}
