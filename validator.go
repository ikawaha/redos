package redos

import (
	"errors"
	"fmt"
	"regexp/syntax"

	"github.com/ikawaha/redos/parser"
)

type RegexSyntax int

func (s RegexSyntax) toSyntaxFlags() syntax.Flags {
	switch s {
	case SyntaxPerl:
		return syntax.Perl
	case SyntaxPOSIX:
		return syntax.POSIX
	default:
		return syntax.Perl // Default to SyntaxPerl if not specified
	}
}

const (
	SyntaxPerl RegexSyntax = iota + 1
	SyntaxPOSIX

	DefaultRegexSyntax     = SyntaxPerl
	DefaultMaxRepeatLimit  = 50
	DefaultStarHeightLimit = 1
	DefaultMaxSizeLimit    = 100000 // 100KB
)

type Validator struct {
	limit  *Complexity
	syntax syntax.Flags
}

type Complexity struct {
	// MaxRepeat is the maximum number of times a subexpression can be repeated.
	MaxRepeat int `json:"max_repeat"`
	// StarHeight is the maximum height of a star in the regex.
	StarHeight int `json:"star_height"`
	// MaxSize is the maximum size of the regex in bytes.
	MaxSize int `json:"max_size"`
}

func NewDefaultLimit() *Complexity {
	return &Complexity{
		MaxRepeat:  DefaultMaxRepeatLimit,
		StarHeight: DefaultStarHeightLimit,
		MaxSize:    DefaultMaxSizeLimit,
	}
}

type Option func(validator *Validator)

func WithLimit(c *Complexity) Option {
	return func(validator *Validator) {
		validator.limit = c
	}
}

func WithSyntax(s RegexSyntax) Option {
	return func(validator *Validator) {
		validator.syntax = s.toSyntaxFlags()
	}
}

func NewValidator(options ...Option) *Validator {
	ret := &Validator{
		limit:  NewDefaultLimit(),
		syntax: DefaultRegexSyntax.toSyntaxFlags(),
	}
	for _, option := range options {
		option(ret)
	}
	return ret
}

func (v Validator) Complexity(re string) (*Complexity, error) {
	var ret Complexity
	if re == "" {
		return &ret, nil
	}
	
	// If all limits are negative, we may skip parsing entirely
	if v.limit.MaxRepeat < 0 && v.limit.StarHeight < 0 && v.limit.MaxSize < 0 {
		ret.MaxRepeat = -1
		ret.StarHeight = -1
		ret.MaxSize = -1
		return &ret, nil
	}
	
	r, err := parser.Parse(re, v.syntax)
	if err != nil {
		return nil, fmt.Errorf("failed to parse regex: %w", err)
	}
	
	// Calculate MaxRepeat only if limit is not negative
	if v.limit.MaxRepeat < 0 {
		ret.MaxRepeat = -1
	} else {
		ret.MaxRepeat = parser.MaxRepeat(r)
	}
	
	// Calculate StarHeight only if limit is not negative
	if v.limit.StarHeight < 0 {
		ret.StarHeight = -1
	} else {
		ret.StarHeight = parser.StarHeight(r)
	}
	
	// Calculate MaxSize only if limit is not negative
	if v.limit.MaxSize < 0 {
		ret.MaxSize = -1
	} else {
		ret.MaxSize = parser.RegexSize(r)
	}
	
	return &ret, nil
}

func (v Validator) Validate(re string) error {
	c, err := v.Complexity(re)
	if err != nil {
		return fmt.Errorf("failed to validate regex: %w", err)
	}
	var errs error
	// Check MaxRepeat only if limit is not negative
	if v.limit.MaxRepeat >= 0 && c.MaxRepeat > v.limit.MaxRepeat {
		errs = errors.Join(errs, fmt.Errorf("regex exceeds max repeat limit: %d > %d", c.MaxRepeat, v.limit.MaxRepeat))
	}
	// Check StarHeight only if limit is not negative
	if v.limit.StarHeight >= 0 && c.StarHeight > v.limit.StarHeight {
		errs = errors.Join(errs, fmt.Errorf("regex exceeds star height limit: %d > %d", c.StarHeight, v.limit.StarHeight))
	}
	// Check MaxSize only if limit is not negative
	if v.limit.MaxSize >= 0 && c.MaxSize > v.limit.MaxSize {
		errs = errors.Join(errs, fmt.Errorf("regex exceeds max size limit: %d > %d", c.MaxSize, v.limit.MaxSize))
	}
	return errs
}
