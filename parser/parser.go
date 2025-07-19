package parser

import (
	"regexp/syntax"
)

// cf. https://golang.org/pkg/regexp/syntax/#Regexp
// A Regexp is a node in a regular expression syntax tree.
// type Regexp struct {
//      Op       Op         // operator
//      Flags    Flags
//      Sub      []*Regexp  // subexpressions, if any
//      Sub0     [1]*Regexp // storage for short Sub
//      Rune     []rune     // matched runes, for OpLiteral, OpCharClass
//      Rune0    [2]rune    // storage for short Rune
//      Min, Max int        // min, max for OpRepeat
//      Cap      int        // capturing index, for OpCapture
//      Name     string     // capturing name, for OpCapture
// }

func Parse(re string, flags syntax.Flags) (*syntax.Regexp, error) {
	return syntax.Parse(re, flags)
}
func StarHeight(r *syntax.Regexp) int {
	switch r.Op {
	case syntax.OpLiteral:
		return 0
	case syntax.OpStar, syntax.OpPlus:
		cost := StarHeight(r.Sub[0])
		return cost + 1
	case syntax.OpRepeat:
		cost := StarHeight(r.Sub[0])
		if r.Max < 0 { // unbounded repeat, like `.*`
			return cost + 1
		}
		return cost
	default:
		var m int
		for _, v := range r.Sub {
			c := StarHeight(v)
			m = max(m, c)
		}
		return m
	}
}

func MaxRepeat(r *syntax.Regexp) int {
	switch r.Op {
	case syntax.OpLiteral:
		return 0
	case syntax.OpRepeat:
		return r.Max
	default:
		var m int
		for _, v := range r.Sub {
			c := MaxRepeat(v)
			m = max(m, c)
		}
		return m
	}
}
