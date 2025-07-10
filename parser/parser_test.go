package parser

import (
	"regexp/syntax"
	"testing"
)

func TestParsePerl(t *testing.T) {
	tests := []struct {
		name     string
		regex    string
		hasError bool
	}{
		{"valid_empty", "", false},
		{"valid_literal", "a", false},
		{"valid_star", "a*", false},
		{"valid_plus", "a+", false},
		{"valid_repeat", "a{1,3}", false},
		{"valid_unbound_repeat", "a{1,}", false},
		{"valid_named_capture", "(?<foo>a|b)*c*", false},
		{"valid_complex", "(a|b)*c+d{2,4}", false},
		{"valid_deeply_nested", "(((a|b)*c)*)*", false},
		{"invalid_unclosed", "(a|b", true},
		{"invalid_invalid_syntax", "a{1,1001}", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Parse(test.regex, syntax.Perl); (err != nil) != test.hasError {
				t.Errorf("Parse(%q) error = %v; want error? %v", test.regex, err, test.hasError)
			}
		})
	}
}

func TestParsePosix(t *testing.T) {
	tests := []struct {
		name     string
		regex    string
		hasError bool
	}{
		{"valid_empty", "", false},
		{"valid_literal", "a", false},
		{"valid_star", "a*", false},
		{"valid_plus", "a+", false},
		{"valid_repeat", "a{1,3}", false},
		{"valid_unbound_repeat", "a{1,}", false},
		{"invalid_named_capture", "(?<foo>a|b)*c*", true},
		{"valid_complex", "(a|b)*c+d{2,4}", false},
		{"valid_deeply_nested", "(((a|b)*c)*)*", false},
		{"invalid_unclosed", "(a|b", true},
		{"invalid_invalid_syntax", "a{1,1001}", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Parse(test.regex, syntax.POSIX); (err != nil) != test.hasError {
				t.Errorf("Parse(%q) error = %v; want error? %v", test.regex, err, test.hasError)
			}
		})
	}
}

func TestStarHeight(t *testing.T) {
	tests := []struct {
		name     string
		regex    string
		expected int
	}{
		{"empty", "", 0},
		{"literal", "a", 0},
		{"star", "a*", 1},
		{"plus", "a+", 1},
		{"repeat", "a{1,3}", 0},
		{"unbound repeat", "a{1,}", 1},
		{"nested", "(a|b)*c*", 1},
		{"complex", "(a|b)*c+d{2,4}", 1},
		{"deeply nested", "(((a|b)*c)*)*", 3},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r, err := Parse(test.regex, syntax.Perl)
			if err != nil {
				t.Fatalf("Failed to parse regex %q: %v", test.regex, err)
			}
			height := StarHeight(r)
			if height != test.expected {
				t.Errorf("StarHeight(%q) = %d; want %d", test.regex, height, test.expected)
			}
		})
	}
}

func TestMaxRepeat(t *testing.T) {
	tests := []struct {
		name     string
		regex    string
		expected int
	}{
		{"empty", "", 0},
		{"literal", "a", 0},
		{"no repeats", "a|b|c", 0},
		{"star", "a*", 0},
		{"plus", "a+", 0},
		{"repeat", "a{1,3}", 3},
		{"nested", "(a|b)*c*", 0},
		{"complex", "(a|b)*c+d{2,4}", 4},
		{"deeply nested", "(((a|b)*c)*)*", 0},
		{"multiple repeats", "a{2,5}b{3,6}c{9}", 9},
		{"unbounded repeat", "a{1,10}b{2,}", 10},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r, err := Parse(test.regex, syntax.Perl)
			if err != nil {
				t.Fatalf("Failed to parse regex %q: %v", test.regex, err)
			}
			max := MaxRepeat(r)
			if max != test.expected {
				t.Errorf("MaxRepeat(%q) = %d; want %d", test.regex, max, test.expected)
			}
		})
	}
}

// TestRealWorldPatterns tests real-world regex patterns based on the documentation examples
func TestRealWorldPatterns(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		starHeight int
		maxRepeat  int
		hasError   bool
	}{
		// pass cases - safe patterns
		{name: "simple_literal", pattern: `^hello world$`, starHeight: 0, maxRepeat: 0},
		{name: "email_validation", pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, starHeight: 1, maxRepeat: 0},
		{name: "phone_number", pattern: `^\+?1?[-\.\s]?\(?[0-9]{3}\)?[-\.\s]?[0-9]{3}[-\.\s]?[0-9]{4}$`, starHeight: 0, maxRepeat: 4},
		{name: "hex_color", pattern: `^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`, starHeight: 0, maxRepeat: 6},
		{name: "date_format", pattern: `^[0-9]{4}-[0-9]{2}-[0-9]{2}$`, starHeight: 0, maxRepeat: 4},
		{name: "username", pattern: `^[a-zA-Z0-9_]{3,20}$`, starHeight: 0, maxRepeat: 20},
		{name: "file_extension", pattern: `\.(jpg|jpeg|png|gif|bmp)$`, starHeight: 0, maxRepeat: 0},
		{name: "identifier_whitelist", pattern: `^[a-zA-Z0-9_]+$`, starHeight: 1, maxRepeat: 0},
		{name: "version_number", pattern: `^v?[0-9]+\.[0-9]+\.[0-9]+$`, starHeight: 1, maxRepeat: 0},
		{name: "html_tag", pattern: `^<[a-zA-Z][a-zA-Z0-9]*>$`, starHeight: 1, maxRepeat: 0},
		{name: "time_format", pattern: `^([01]?[0-9]|2[0-3]):[0-5][0-9]$`, starHeight: 0, maxRepeat: 0},
		{name: "boundary_repeat", pattern: `^a{1,1000}$`, starHeight: 0, maxRepeat: 1000},
		{name: "complex_alternation", pattern: `^(apple|banana|cherry|date|elderberry)$`, starHeight: 0, maxRepeat: 0},
		{name: "word_boundary", pattern: `\b[a-zA-Z]+\b`, starHeight: 1, maxRepeat: 0},

		// fail cases - ReDoS patterns
		{name: "nested_quantifiers_classic", pattern: `^(a+)+$`, starHeight: 2, maxRepeat: 0},
		{name: "nested_quantifiers_with_suffix", pattern: `^(a+)+b$`, starHeight: 2, maxRepeat: 0},
		{name: "multiple_quantifiers_nested", pattern: `^(a*b*)*$`, starHeight: 2, maxRepeat: 0},
		{name: "triple_nesting", pattern: `^((a+)*b*)*$`, starHeight: 3, maxRepeat: 0},
		{name: "quantified_alternation", pattern: `^(a+|b+)*$`, starHeight: 2, maxRepeat: 0},
		{name: "outer_quantifier_on_group", pattern: `^(ab+)+$`, starHeight: 2, maxRepeat: 0},
		{name: "high_repeat_bound", pattern: `^a{1,1000}$`, starHeight: 0, maxRepeat: 1000},
		{name: "boundary_plus_one", pattern: `^a{1,1001}$`, starHeight: 0, maxRepeat: 1001, hasError: true},
		{name: "nested_whitespace_quantifiers", pattern: `^\s*(\S+\s*)*$`, starHeight: 2, maxRepeat: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r, err := Parse(test.pattern, syntax.Perl)
			if err != nil {
				t.Fatalf("Failed to parse regex %q: %v", test.pattern, err)
			}

			starHeight := StarHeight(r)
			if starHeight != test.starHeight {
				t.Errorf("StarHeight(%q) = %d; want %d", test.pattern, starHeight, test.starHeight)
			}

			maxRepeat := MaxRepeat(r)
			if maxRepeat != test.maxRepeat {
				t.Errorf("MaxRepeat(%q) = %d; want %d", test.pattern, maxRepeat, test.maxRepeat)
			}
		})
	}
}
