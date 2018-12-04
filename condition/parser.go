package condition

import (
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/servicesdk/wfapi"
	"regexp"
	"strings"
	"text/scanner"
)

var namePattern = regexp.MustCompile(`\A[a-z][a-zA-Z0-9_]*\z`)

type parser struct {
	str string
	scn scanner.Scanner
}

func Parse(str string) wfapi.Condition {
	p := &parser{}
	p.str = str
	p.scn.Init(strings.NewReader(str))
	c, r := p.parseOr()
	if r != scanner.EOF {
		panic(eval.Error(WF_CONDITION_SYNTAX_ERROR, issue.H{`text`: p.str, `pos`: p.scn.Offset}))
	}
	return c
}

func (p *parser) parseOr() (wfapi.Condition, rune) {
	es := make([]wfapi.Condition, 0)
	for {
		lh, r := p.parseAnd()
		es = append(es, lh)
		if p.scn.TokenText() != `or` {
			if len(es) == 1 {
				return es[0], r
			}
			return newOr(es), r
		}
	}
}

func (p *parser) parseAnd() (wfapi.Condition, rune) {
	es := make([]wfapi.Condition, 0)
	for {
		lh, r := p.parseUnary()
		es = append(es, lh)
		if p.scn.TokenText() != `and` {
			if len(es) == 1 {
				return es[0], r
			}
			return newAnd(es), r
		}
	}
}

func (p *parser) parseUnary() (c wfapi.Condition, r rune) {
	r = p.scn.Scan()
	if r == '!' {
		c, r = p.parseAtom(p.scn.Scan())
		return newNot(c), r
	}
	return p.parseAtom(r)
}

func (p *parser) parseAtom(r rune) (wfapi.Condition, rune) {
	if r == scanner.EOF {
		panic(eval.Error(WF_CONDITION_UNEXPECTED_END, issue.H{`text`: p.str, `pos`: p.scn.Offset}))
	}

	if r == '(' {
		var c wfapi.Condition
		c, r = p.parseOr()
		if r != ')' {
			panic(eval.Error(WF_CONDITION_MISSING_RP, issue.H{`text`: p.str, `pos`: p.scn.Offset}))
		}
		return c, p.scn.Scan()
	}
	w := p.scn.TokenText()
	if namePattern.MatchString(w) {
		r = p.scn.Scan()
		switch w {
		case `true`:
			return Always, r
		case `false`:
			return Never, r
		default:
			return newTruthy(w), r
		}
	}
	panic(eval.Error(WF_CONDITION_INVALID_NAME, issue.H{`name`: w, `text`: p.str, `pos`: p.scn.Offset}))
}
