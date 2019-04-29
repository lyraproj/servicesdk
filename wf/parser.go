package wf

import (
	"regexp"
	"strings"
	"text/scanner"

	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
)

var namePattern = regexp.MustCompile(`\A[a-z][a-zA-Z0-9_]*\z`)

type parser struct {
	str string
	scn scanner.Scanner
}

func Parse(str string) Condition {
	if str == `` {
		return Always
	}
	p := &parser{}
	p.str = str
	p.scn.Init(strings.NewReader(str))
	c, r := p.parseOr()
	if r != scanner.EOF {
		panic(px.Error(ConditionSyntaxError, issue.H{`text`: p.str, `pos`: p.scn.Offset}))
	}
	return c
}

func (p *parser) parseOr() (Condition, rune) {
	es := make([]Condition, 0)
	for {
		lh, r := p.parseAnd()
		es = append(es, lh)
		if p.scn.TokenText() != `or` {
			if len(es) == 1 {
				return es[0], r
			}
			return Or(es), r
		}
	}
}

func (p *parser) parseAnd() (Condition, rune) {
	es := make([]Condition, 0)
	for {
		lh, r := p.parseUnary()
		es = append(es, lh)
		if p.scn.TokenText() != `and` {
			if len(es) == 1 {
				return es[0], r
			}
			return And(es), r
		}
	}
}

func (p *parser) parseUnary() (c Condition, r rune) {
	r = p.scn.Scan()
	if r == '!' {
		c, r = p.parseAtom(p.scn.Scan())
		return Not(c), r
	}
	return p.parseAtom(r)
}

func (p *parser) parseAtom(r rune) (Condition, rune) {
	if r == scanner.EOF {
		panic(px.Error(ConditionUnexpectedEnd, issue.H{`text`: p.str, `pos`: p.scn.Offset}))
	}

	if r == '(' {
		var c Condition
		c, r = p.parseOr()
		if r != ')' {
			panic(px.Error(ConditionMissingRp, issue.H{`text`: p.str, `pos`: p.scn.Offset}))
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
			return Truthy(w), r
		}
	}
	panic(px.Error(ConditionInvalidName, issue.H{`name`: w, `text`: p.str, `pos`: p.scn.Offset}))
}
